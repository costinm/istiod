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
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/costinm/grpc-mesh/echo-micro/server"
	"google.golang.org/grpc/grpclog"

	//"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"github.com/costinm/grpc-mesh/telemetry/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/host"

	"go.opentelemetry.io/contrib/zpages"
	"go.opentelemetry.io/otel/exporters/prometheus"

	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	metric_controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	metric_processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

// TODO: use otelhttptrace to get httptrace (low level client traces)

// OTel doesn't have metrics in grpc instrumentation.
// When they do, we can switch from OpenCensus.

//func HttpRoundTripper() func(transport http.RoundTripper) http.RoundTripper {
//	return func(transport http.RoundTripper) http.RoundTripper {
//		return otelhttp.NewTransport(transport)
//	}
//}

var (
	log = grpclog.Component("uecho")
)

func initTracing(ctx context.Context, r *resource.Resource) func() {
	// For traces, zpages seems better for debugging.
	// Additional otel exporter can be added.
	//exp, err = stdouttrace.New(
	//	stdouttrace.WithWriter(os.Stderr),
	//	// Use human readable output.
	//	stdouttrace.WithPrettyPrint(),
	//	// Do not print timestamps for the demo.
	//	//stdouttrace.WithoutTimestamps(),
	//)
	//if err != nil {
	//	return nil, err
	//}

	// Span processor for tracez
	sp := zpages.NewSpanProcessor()
	thandler := zpages.NewTracezHandler(sp)
	http.DefaultServeMux.Handle("/debug/tracez", thandler)

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(r),
		//trace.WithBatcher(exp),
		trace.WithSpanProcessor(sp),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}
}

func initOTel(ctx context.Context, serviceName string) (func(), error) {
	r := resource.NewWithAttributes(semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName))

	tracingCleanup := initTracing(ctx, r)

	// =========== Metrics

	// Where to send the metrics
	//exporter, err := stdoutmetric.New(
	//	stdoutmetric.WithPrettyPrint(),
	//	stdoutmetric.WithWriter(os.Stderr))
	//if err != nil {
	//	log.Fatalf("creating stdoutmetric exporter: %v", err)
	//}

	// Push controller for metrics
	metricsController := metric_controller.New(
		metric_processor.NewFactory(
			simple.NewWithInexpensiveDistribution(),
			aggregation.CumulativeTemporalitySelector(), //exporter,
			metric_processor.WithMemory(true),
		),
		//metric_controller.WithExporter(exporter),
		// default: 10s
		// for push - interval between metric pushes
		// for prom: do not recompute more often
		//metric_controller.WithCollectPeriod(60*time.Second),
		metric_controller.WithResource(r),
		// WithResource, WithPushTimeout
	)

	config := prometheus.Config{
		DefaultHistogramBoundaries: []float64{1, 2, 5, 10, 20, 50},
	}
	exporterp, err := prometheus.New(config, metricsController)
	if err != nil {
		return nil, err
	}
	http.HandleFunc("/metrics", exporterp.ServeHTTP)

	//if err = metricsController.RoundTripStart(ctx); err != nil {
	//	log.Fatalf("starting push controller: %v", err)
	//}

	global.SetMeterProvider(metricsController)

	// Global instrumentations
	//if err := runtime.RoundTripStart(
	//	runtime.WithMinimumReadMemStatsInterval(time.Second),
	//); err != nil {
	//	log.Fatalln("failed to start runtime instrumentation:", err)
	//}
	// Host telemetry -
	host.Start()

	// End telemetry magic
	return func() {
		if err := metricsController.Stop(ctx); err != nil {
			log.Fatalf("stopping push controller: %v", err)
		}
		tracingCleanup()
	}, nil
}

func Run(lis net.Listener) error {
	h := &server.EchoGrpcHandler{
		// Enable OpenTelemetry client side
		DialOptions: []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		},
	}

	creds := insecure.NewCredentials()

	grpcOptions := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
	}

	grpcServer := grpc.NewServer(grpcOptions...)
	h.Register(grpcServer)
	reflection.Register(grpcServer)

	go grpcServer.Serve(lis)

	return nil
}

// Most minimal gRPC based server, for estimating binary size overhead.
//
// - 0.8M for a min go program
// - 4.7M for an echo using HTTP.
// - 9M - this server, only plain gRPC
// - 20M - same app, but proxyless gRPC
// - 22M - plus opencensus, prom, zpages, reflection
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9280"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	initOTel(context.Background(), "echo")

	err = Run(lis)
	if err != nil {
		fmt.Println("Error ", err)
	}
	go http.ListenAndServe("127.0.0.1:9281", http.DefaultServeMux)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

}
