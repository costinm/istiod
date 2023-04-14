// Copyright Istio Authors
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

package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	ocprom "contrib.go.opencensus.io/exporter/prometheus"
	"github.com/GoogleCloudPlatform/cloud-run-mesh/pkg/k8s"
	"github.com/GoogleCloudPlatform/cloud-run-mesh/pkg/mesh"
	"github.com/costinm/grpc-mesh/bootstrap"
	"github.com/costinm/grpc-mesh/echo-micro/server"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/runmetrics"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/zpages"
	"google.golang.org/grpc/binarylog"
	pb "google.golang.org/grpc/binarylog/grpc_binarylog_v1"

	//grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_zap "github.com/costinm/grpc-mesh/telemetry/logs/zap"
	"go.uber.org/zap"

	// Instrumentations
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/admin"
	"google.golang.org/grpc/credentials/insecure"
	xdscreds "google.golang.org/grpc/credentials/xds"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/health"
	grpcHealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/xds"
)

// Istio echo server with:
// - telemetry/traces (prom/zpages)
// - XDS
// - zap logs
//

var log = grpclog.Component("echo")

// GRPCServer is the interface implemented by both grpc
type GRPCServer interface {
	RegisterService(*grpc.ServiceDesc, interface{})
	Serve(net.Listener) error
	Stop()
	GracefulStop()
	GetServiceInfo() map[string]grpc.ServiceInfo
}

// TODO: 2 servers, one XDS and one plain
// TODO: get certs, remote config, JWTs
// TODO: tunnels
func Run(lis net.Listener) (func(), error) {
	// Configure zap as a logger for grpc.
	zl, _ := zap.NewDevelopment(zap.AddCallerSkip(4))
	if strings.Contains(os.Getenv("DEBUG"), "xds") {
		grpc_zap.ReplaceGrpcLoggerV2WithVerbosity(zl, -2)
	}

	alwaysLoggingDeciderServer := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return true }
	alwaysLoggingDeciderClient := func(ctx context.Context, fullMethodName string) bool { return true }

	// Init metrics and telemetry
	cleanup, err := initTel(context.Background(), "echo")
	if err != nil {
		return nil, err
	}

	grpcOptions := []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			otelgrpc.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(zl),
			grpc_zap.PayloadStreamServerInterceptor(zl, alwaysLoggingDeciderServer)),
		grpc_middleware.WithUnaryServerChain(
			otelgrpc.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(zl),
			grpc_zap.PayloadUnaryServerInterceptor(zl, alwaysLoggingDeciderServer)),
	}

	var grpcServer GRPCServer
	bf := os.Getenv("GRPC_XDS_BOOTSTRAP")
	if bf != "" {
		creds, _ := xdscreds.NewServerCredentials(xdscreds.ServerOptions{
			FallbackCreds: insecure.NewCredentials()})
		grpcOptions = append(grpcOptions, grpc.Creds(creds))

		if _, err := os.Stat(bf); os.IsNotExist(err) || true {
			// Only needs to be set to a file - if the file doesn't exist, create it.
			kr := mesh.New()

			// Get mesh config from the current cluster
			// Using a K8s Client
			kc := &k8s.K8S{Mesh: kr}
			err := kc.K8SClient(context.Background())
			if err != nil {
				return nil, err
			}
			kr.Cfg = kc
			kr.TokenProvider = kc

			// Load the config map with mesh-env
			kr.LoadConfig(context.Background())

			if kr.XDSAddr == "" {
				kr.XDSAddr = kr.MeshConnectorAddr + ":15012"
			}

			err = bootstrap.GenerateBootstrapFile(&bootstrap.GenerateBootstrapOptions{
				DiscoveryAddress: kr.XDSAddr,
			}, bf)
			if err != nil {
				log.Fatal("Failed to write bootstrap file", err)
			}
			log.Info("Auto-generated bootstrap ", bf)
		} else {
			// Istio injected creds and bootstrap
			log.Info("Using existing bootstrap ", bf)
		}
		grpcServer = xds.NewGRPCServer(grpcOptions...)
	} else {
		// Istio sidecar mode.
		grpcServer = grpc.NewServer(grpcOptions...)
	}

	// Special handling for startup without env variable set

	// Generate the bootstrap if the file is missing ( injection-less )
	// using cloudrun-mesh auto-detection code

	// TODO: Generate certs if missing
	// Configure the client side options.
	h := &server.EchoGrpcHandler{
		// Enable OpenTelemetry client side
		DialOptions: []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
			grpc.WithUnaryInterceptor(grpc_zap.UnaryClientInterceptor(zl)),
			grpc.WithStreamInterceptor(grpc_zap.StreamClientInterceptor(zl)),
			grpc.WithUnaryInterceptor(grpc_zap.PayloadUnaryClientInterceptor(zl, alwaysLoggingDeciderClient)),
			grpc.WithStreamInterceptor(grpc_zap.PayloadStreamClientInterceptor(zl, alwaysLoggingDeciderClient)),
		},
	}

	h.Register(grpcServer)

	http.Handle("/grpc/", h)

	// add the standard grpc health check
	healthServer := health.NewServer()

	grpcHealth.RegisterHealthServer(grpcServer, healthServer)
	reflection.Register(grpcServer)

	// Log all requests. This is a global setting, captures all requests.
	initBinlog()

	// grpcdebug support
	_, err = admin.Register(grpcServer)
	if err != nil {
		log.Info("Failed to register admin", "error", err)
	}

	// Status
	go http.ListenAndServe(":9090", http.DefaultServeMux)

	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			log.Fatal(err)
		}
	}()
	return cleanup, nil
}

type DebugBinaryLogSink struct {
}

func (d DebugBinaryLogSink) Write(entry *pb.GrpcLogEntry) error {
	log.Info("binarylog", "proto", entry.String())
	return nil
}

func (d DebugBinaryLogSink) Close() error {
	return nil
}

// Debug/test: set binary log
func initBinlog() {
	binarylog.SetSink(&DebugBinaryLogSink{})
}

func initTel(background context.Context, s string) (func(), error) {
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		log.Errorf("Failed to register ocgrpc server views: %v", err)
	}
	if err := view.Register(ocgrpc.DefaultClientViews...); err != nil {
		log.Errorf("Failed to register ocgrpc server views: %v", err)
	}

	// Similar with pilot-agent - init prometheus registry, add a prefix
	registry := prometheus.NewRegistry()

	//wrapped := prometheus.WrapRegistererWithPrefix("uecho_", prometheus.Registerer(registry))

	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())

	// go collector metrics collide with other metrics.
	exporter, err := ocprom.NewExporter(ocprom.Options{
		Registry: registry})
	//Registerer: wrapped})
	if err != nil {
		log.Fatalf("could not setup exporter: %v", err)
	}

	view.RegisterExporter(exporter)

	zpages.Handle(http.DefaultServeMux, "/debug")

	http.Handle("/metrics", exporter)

	// Collect runtime metrics
	err = runmetrics.Enable(runmetrics.RunMetricOptions{
		EnableCPU:    true,
		EnableMemory: true,
		Prefix:       "echo/",
	})
	if err != nil {
		log.Error(err)
	}

	return func() {

	}, nil
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "17071"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	c, err := Run(lis)
	if err != nil {
		log.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	c()
	// TODO: lame duck, etc
}
