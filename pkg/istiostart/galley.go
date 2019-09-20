// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package istiostart

import (
	"fmt"
	"istio.io/istio/galley/pkg/config/meshcfg"
	"istio.io/istio/galley/pkg/config/processor"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	grpcMetadata "google.golang.org/grpc/metadata"

	mcp "istio.io/api/mcp/v1alpha1"
	fs2 "istio.io/istio/galley/pkg/config/source/kube/fs"

	"istio.io/pkg/ctrlz/fw"
	"istio.io/pkg/log"
	"istio.io/pkg/version"

	"istio.io/istio/galley/pkg/config/event"
	"istio.io/istio/galley/pkg/config/processing"
	"istio.io/istio/galley/pkg/config/processing/snapshotter"
	"istio.io/istio/galley/pkg/config/processor/metadata"
	"istio.io/istio/galley/pkg/config/processor/transforms"
	"istio.io/istio/galley/pkg/config/schema"
	"istio.io/istio/galley/pkg/config/source/kube"
	"istio.io/istio/galley/pkg/config/source/kube/rt"
	"istio.io/istio/galley/pkg/runtime/groups"
	"istio.io/istio/galley/pkg/server/process"
	"istio.io/istio/galley/pkg/server/settings"
	configz "istio.io/istio/pkg/mcp/configz/server"
	"istio.io/istio/pkg/mcp/monitoring"
	mcprate "istio.io/istio/pkg/mcp/rate"
	"istio.io/istio/pkg/mcp/server"
	"istio.io/istio/pkg/mcp/snapshot"
	"istio.io/istio/pkg/mcp/source"
)

// Processing2 component is the main config processing component that will listen to a config source and publish
// resources through an MCP server.

// This is a simplified startup for galley, specific for hyperistio/combined:
// - callout removed - standalone galley supports it, and should be used
// - acl removed - envoy and Istio RBAC should handle it
// - listener removed - common grpc server for all components, using Pilot's listener


type Processing2 struct {
	args *settings.Args

	mcpCache     *snapshot.Cache
	configzTopic fw.Topic

	k kube.Interfaces

	serveWG       sync.WaitGroup
	grpcServer    *grpc.Server
	runtime       *processing.Runtime
	mcpSource     *source.Server
	reporter      monitoring.Reporter
	listenerMutex sync.Mutex
	listener      net.Listener
	stopCh        chan struct{}
}

var scope = log.RegisterScope("server", "", 0)
var _ process.Component = &Processing2{}

const versionMetadataKey = "config.source.version"

// NewProcessing2 returns a new processing component.
func NewProcessing2(a *settings.Args) *Processing2 {
	mcpCache := snapshot.New(groups.IndexFunction)
	return &Processing2{
		args:         a,
		mcpCache:     mcpCache,
		configzTopic: configz.CreateTopic(mcpCache),
	}
}

// Start implements process.Component
func (p *Processing2) Start() (err error) {
	var mesh event.Source
	var src event.Source

	// This returns the default mesh config, without a way to override
	if mesh, err = meshcfg.NewFS(p.args.MeshConfigFile); err != nil {
		return
	}

	m := metadata.MustGet()

	kubeResources := p.disableExcludedKubeResources(m)

	if src, err = p.createSource(kubeResources); err != nil {
		return
	}

	var distributor snapshotter.Distributor = snapshotter.NewMCPDistributor(p.mcpCache)
	transformProviders := transforms.Providers(m)

	if p.runtime, err = processor.Initialize(m, p.args.DomainSuffix, event.CombineSources(mesh, src), transformProviders, distributor); err != nil {
		return
	}

	grpcOptions := p.getServerGrpcOptions()

	p.stopCh = make(chan struct{})
	var checker source.AuthChecker = server.NewAllowAllChecker()

	grpc.EnableTracing = p.args.EnableGRPCTracing
	p.grpcServer = grpc.NewServer(grpcOptions...)

	p.reporter = monitoring.NewStatsContext("galley")

	options := &source.Options{
		Watcher:            p.mcpCache,
		Reporter:           p.reporter,
		CollectionsOptions: source.CollectionOptionsFromSlice(m.AllCollectionsInSnapshots()),
		ConnRateLimiter:    mcprate.NewRateLimiter(time.Second, 100), // TODO(Nino-K): https://github.com/istio/istio/issues/12074
	}

	// TODO: reimplement callout ( SinkAddress ), or leave it to standalone galley

	md := grpcMetadata.MD{
		versionMetadataKey: []string{version.Info.Version},
	}
	serverOptions := &source.ServerOptions{
		AuthChecker: checker,
		RateLimiter: rate.NewLimiter(rate.Every(time.Second), 100), // TODO(Nino-K): https://github.com/istio/istio/issues/12074
		Metadata:    md,
	}

	p.mcpSource = source.NewServer(options, serverOptions)

	// get the network stuff setup
	network := "tcp"
	var address string
	idx := strings.Index(p.args.APIAddress, "://")
	if idx < 0 {
		address = p.args.APIAddress
	} else {
		network = p.args.APIAddress[:idx]
		address = p.args.APIAddress[idx+3:]
	}

	if p.listener, err = net.Listen(network, address); err != nil {
		err = fmt.Errorf("unable to listen: %v", err)
		return
	}

	mcp.RegisterResourceSourceServer(p.grpcServer, p.mcpSource)

	var startWG sync.WaitGroup
	startWG.Add(1)

	p.serveWG.Add(1)
	go func() {
		defer p.serveWG.Done()
		p.runtime.Start()

		l := p.getListener()
		if l != nil {
			// start serving
			gs := p.grpcServer
			startWG.Done()
			err = gs.Serve(l)
			if err != nil {
				scope.Errorf("Galley Server unexpectedly terminated: %v", err)
			}
		}
	}()

	startWG.Wait()

	return nil
}

func (p *Processing2) disableExcludedKubeResources(m *schema.Metadata) schema.KubeResources {

	// Behave in the same way as existing logic:
	// - Builtin types are excluded by default.
	// - If ServiceDiscovery is enabled, any built-in type should be readded.

	var result schema.KubeResources
	for _, r := range m.KubeSource().Resources() {

		if p.isKindExcluded(r.Kind) {
			// Found a matching exclude directive for this KubeResource. Disable the resource.
			r.Disabled = true

			// Check and see if this is needed for Service Discovery. If needed, we will need to re-enable.
			if p.args.EnableServiceDiscovery {
				// IsBuiltIn is a proxy for types needed for service discovery
				a := rt.DefaultProvider().GetAdapter(r)
				if a.IsBuiltIn() {
					// This is needed for service discovery. Re-enable.
					r.Disabled = false
				}
			}
		}

		result = append(result, r)
	}

	return result
}

// ConfigZTopic returns the ConfigZTopic for the processor.
func (p *Processing2) ConfigZTopic() fw.Topic {
	return p.configzTopic
}

func (p *Processing2) getServerGrpcOptions() []grpc.ServerOption {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(uint32(p.args.MaxConcurrentStreams)),
		grpc.MaxRecvMsgSize(int(p.args.MaxReceivedMessageSize)),
		grpc.InitialWindowSize(int32(p.args.InitialWindowSize)),
		grpc.InitialConnWindowSize(int32(p.args.InitialConnectionWindowSize)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Timeout:               p.args.KeepAlive.Timeout,
			Time:                  p.args.KeepAlive.Time,
			MaxConnectionAge:      p.args.KeepAlive.MaxServerConnectionAge,
			MaxConnectionAgeGrace: p.args.KeepAlive.MaxServerConnectionAgeGrace,
		}),
		// Relax keepalive enforcement policy requirements to avoid dropping connections due to too many pings.
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	return grpcOptions
}

// Abstract the factory to create the config source
type GalleyCfgSourceFn func(resources schema.KubeResources) (src event.Source, err error)

func (p *Processing2) createSource(resources schema.KubeResources) (src event.Source, err error) {
		if src, err = fs2.New(p.args.ConfigPath, resources); err != nil {
			return
		}
		return
}

func (p *Processing2) isKindExcluded(kind string) bool {
	for _, excludedKind := range p.args.ExcludedResourceKinds {
		if kind == excludedKind {
			return true
		}
	}

	return false
}

// Stop implements process.Component
func (p *Processing2) Stop() {
	if p.stopCh != nil {
		close(p.stopCh)
		p.stopCh = nil
	}

	if p.grpcServer != nil {
		p.grpcServer.GracefulStop()
		p.grpcServer = nil
	}

	if p.runtime != nil {
		p.runtime.Stop()
		p.runtime = nil
	}

	p.listenerMutex.Lock()
	if p.listener != nil {
		_ = p.listener.Close()
		p.listener = nil
	}
	p.listenerMutex.Unlock()

	if p.reporter != nil {
		_ = p.reporter.Close()
		p.reporter = nil
	}

	// final attempt to purge buffered logs
	_ = log.Sync()
}

func (p *Processing2) getListener() net.Listener {
	p.listenerMutex.Lock()
	defer p.listenerMutex.Unlock()
	return p.listener
}

// Address returns the Address of the MCP service.
func (p *Processing2) Address() net.Addr {
	l := p.getListener()
	if l == nil {
		return nil
	}
	return l.Addr()
}
