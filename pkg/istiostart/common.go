package istiostart

import (
	"fmt"
	"github.com/gogo/protobuf/types"
	"istio.io/istio/galley/pkg/server/settings"
	"istio.io/istio/pilot/pkg/proxy/envoy"
	"istio.io/istio/pkg/config/mesh"
	"istio.io/istio/pkg/keepalive"
	"istio.io/istio/security/pkg/nodeagent/cache"
	"istio.io/istio/security/pkg/nodeagent/sds"
	"istio.io/istio/security/pkg/nodeagent/secretfetcher"
	"istio.io/pkg/ctrlz"
	"istio.io/pkg/filewatcher"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	meshv1 "istio.io/api/mesh/v1alpha1"

	agent "istio.io/istio/pkg/bootstrap"
)

var (
	fileWatcher filewatcher.FileWatcher
)

func (s *Server) InitCommon(args *PilotArgs) {

	if args.CtrlZOptions != nil {
		_, _ = ctrlz.Run(args.CtrlZOptions, nil)
	}

	_, addr, err := startMonitor(args.DiscoveryOptions.MonitoringAddr, s.mux)
	if err != nil {
		return
	}
	s.MonitorListeningAddr = addr

	//go func() {
	//	<-s.stop
	//	err := monitor.Close()
	//	log.Debugf("Monitoring server terminated: %v", err)
	//}()

}

// Start all components of istio, using local config files.
//
// A minimal set of Istio Env variables are also used.
// This is expected to run in a Docker or K8S environment, with a volume with user configs mounted.
//
//
// Defaults:
// - http port 15007
// - grpc on 15010
//- config from $ISTIO_CONFIG or ./conf
func Init() (*Server, error) {
	baseDir := "."
	//meshConfigFile := baseDir + "/conf/pilot/mesh.yaml"

	mcfgObj := mesh.DefaultMeshConfig()

	mcfg := &mcfgObj
	mcfg.AuthPolicy = meshv1.MeshConfig_NONE

	mcfg.DisablePolicyChecks = true
	mcfg.ProxyHttpPort = 12080
	mcfg.ProxyListenPort = 12001

	// TODO: 15006 can't be configured currently
	// TODO: 15090 (prometheus) can't be configured. It's in the bootstrap file, so easy to replace

	mcfg.ProxyHttpPort = 12002
	mcfg.DefaultConfig = &meshv1.ProxyConfig{
		DiscoveryAddress:       "localhost:12010",
		ControlPlaneAuthPolicy: meshv1.AuthenticationPolicy_NONE,

		ProxyAdminPort: 12000,

		ConfigPath: baseDir + "/run",
		// BinaryPath:       "/usr/local/bin/envoy", - default
		CustomConfigFile:       baseDir + "/conf/sidecar/envoy_bootstrap_v2.json",
		ConnectTimeout:         types.DurationProto(5 * time.Second),  // crash if not set
		DrainDuration:          types.DurationProto(30 * time.Second), // crash if 0
		StatNameLength:         189,
		ParentShutdownDuration: types.DurationProto(5 * time.Second),

		ServiceCluster: "istio",
	}

	// Create a test pilot discovery service configured to watch the tempDir.
	args := PilotArgs{
		Namespace: "testing",
		DiscoveryOptions: envoy.DiscoveryServiceOptions{
			HTTPAddr:        ":12007",
			GrpcAddr:        ":12010",
			SecureGrpcAddr:  ":12011",
			EnableCaching:   true,
			EnableProfiling: true,
		},

		Mesh: MeshArgs{

			MixerAddress:    "localhost:9091",
			RdsRefreshDelay: types.DurationProto(10 * time.Millisecond),
		},
		Config: ConfigArgs{},

		// MCP is messing up with the grpc settings...
		MCPMaxMessageSize:        1024 * 1024 * 64,
		MCPInitialWindowSize:     1024 * 1024 * 64,
		MCPInitialConnWindowSize: 1024 * 1024 * 64,

		MeshConfig:       mcfg,
		KeepaliveOptions: keepalive.DefaultOption(),
	}

	// Load config from the in-process Galley.
	// We can also configure Envoy to listen on 9901 and galley on different port, and LB
	mcfg.ConfigSources = []*meshv1.ConfigSource{
		&meshv1.ConfigSource{
			Address: "localhost:12901",
		},
	}

	gargs := settings.DefaultArgs()

	// Default dir.
	// If not set, will attempt to use K8S.
	gargs.ConfigPath = baseDir + "/conf/istio/simple"
	// TODO: load a json file to override defaults (for all components)

	gargs.ValidationArgs.EnableValidation = false
	gargs.ValidationArgs.EnableReconcileWebhookConfiguration = false
	gargs.APIAddress = "tcp://0.0.0.0:12901"
	gargs.Insecure = true
	gargs.EnableServer = true
	gargs.DisableResourceReadyCheck = true
	// Use Galley Ctrlz for all services.
	gargs.IntrospectionOptions.Port = 12876

	// The file is loaded and watched by Galley using galley/pkg/meshconfig watcher/reader
	// Current code in galley doesn't expose it - we'll use 2 Caches instead.

	// Defaults are from pkg/config/mesh

	// Actual files are loaded by galley/pkg/src/fs, which recursively loads .yaml and .yml files
	// The files are suing YAMLToJSON, but interpret Kind, APIVersion

	gargs.MeshConfigFile = baseDir + "/conf/pilot/mesh.yaml"
	gargs.MonitoringPort = 12015

	server, err := NewServer(args)
	if err != nil {
		return nil, err
	}



	server.Galley = NewProcessing2(gargs)

	err = server.Init()
	if err != nil {
		return nil, err
	}

	// Start the SDS server for TLS certs
	err = StartSDS(baseDir, args.MeshConfig)
	if err != nil {
		return nil, err
	}

	// TODO: start envoy only if TLS certs exist (or bootstrap token and SDS server address is configured)
	//err = startEnvoy(baseDir, &mcfg)
	//if err != nil {
	//	return err
	//}
	return server, nil
}

func (s *Server) WaitDrain(baseDir string) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	// Will gradually terminate connections to Pilot
	DrainEnvoy(baseDir, s.Args.MeshConfig.DefaultConfig)

}

// Start the SDS service. Uses the main Istio address.
//
func StartSDS(baseDir string, config *meshv1.MeshConfig) error {

	return nil
}

func StartSDSK8S(baseDir string, config *meshv1.MeshConfig) error {

	// This won't work on VM - only on K8S.
	var workloadSdsCacheOptions cache.Options
	var serverOptions sds.Options

	// Compat with Istio env
	caProvider := os.Getenv("CA_PROVIDER")
	if caProvider == "" {
		caProvider = "Citadel"
	}

	wSecretFetcher, err := secretfetcher.NewSecretFetcher(false,
		serverOptions.CAEndpoint, caProvider, true,
		[]byte(""), "", "", "", "")
	if err != nil {
		log.Fatal("failed to create secretFetcher for workload proxy", err)
	}
	workloadSdsCacheOptions.TrustDomain = serverOptions.TrustDomain
	workloadSdsCacheOptions.Plugins = sds.NewPlugins(serverOptions.PluginNames)
	workloadSecretCache := cache.NewSecretCache(wSecretFetcher, sds.NotifyProxy, workloadSdsCacheOptions)

	// GatewaySecretCache loads secrets from K8S
	_, err = sds.NewServer(serverOptions, workloadSecretCache, nil)

	if err != nil {
		log.Fatal("Failed to start SDS server", err)
	}

	return nil
}

var trustDomain = "cluster.local"

// TODO: use pilot-agent code, and refactor it to extract the core functionality.

// TODO: better implementation for 'drainFile' config - used by agent.terminate()

// startEnvoy starts the envoy sidecar for Istio control plane, for TLS and load balancing.
// Not used otherwise.
func startEnvoy(baseDir string, mcfg *meshv1.MeshConfig) error {
	os.Mkdir(baseDir+"/run", 0700)
	cfg := mcfg.DefaultConfig

	nodeId := "sidecar~127.0.0.2~istio-control.fortio~fortio.svc.cluster.local"
	env := os.Environ()
	env = append(env, "ISTIO_META_ISTIO_VERSION=1.4")

	cfgF, err := agent.WriteBootstrap(cfg, nodeId, 1, []string{
		"istio-pilot.istio-system",
		fmt.Sprintf("spiffe://%s/ns/%s/sa/%s", trustDomain, "istio-system", "istio-pilot-service-account"),
	},
		map[string]interface{}{},
		env,
		[]string{"127.0.0.2"}, // node IPs
		"60s")
	if err != nil {
		return err
	}

	// Start Envoy, using the pre-generated config. No restarts: if it crashes, we exit.
	stop := make(chan error)
	//features.EnvoyBaseId.DefaultValue = "1"
	process, err := agent.RunProxy(cfg, nodeId, 1, cfgF, stop,
		os.Stdout, os.Stderr, []string{
			"--disable-hot-restart",
			// "-l", "trace",
		})
	go func() {
		// Should not happen.
		process.Wait()
		log.Fatal("Envoy terminated, restart.")
	}()
	return err
}

// Start the galley component, with its args.

	// Galley by default initializes some probes - we'll use Pilot probes instead, since it also checks for galley
	// TODO: have  the probes check all other components

	// Validation is not included in hyperistio - standalone Galley or external address should be used, it's not
	// part of the minimal set.

	// Monitoring, profiling are common across all components, skipping as part of Galley startup
	// Ctrlz is custom for galley - setting it up here.


