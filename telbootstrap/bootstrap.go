package telbootstrap

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	//"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	//"github.com/costinm/istiod/telemetry/otelgrpc"

	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	prom "github.com/prometheus/client_golang/prometheus"

	"go.opentelemetry.io/contrib/zpages"
	"go.opentelemetry.io/otel/exporters/prometheus"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	//metric_controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	//"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	//metric_processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	//"go.opentelemetry.io/otel/sdk/metric/selector/simple"

	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
)

// Bootstrap file for open telemetry, to avoid repeating it in each main.
// OTel has a strange organization and versioning:
// Core packages ( used by libraries ) go.opentelemetry.io/otel - tag man be 1.15.1
// But metrics-related packages are v0.38.1
// Contrib has different version - can be 1.16.0

func InitTracing(ctx context.Context, r *resource.Resource) func() {
	// For traces, zpages seems better for debugging - but we can also send to stdout
	//
	// Additional otel exporter can be added.
	stdoutTrace, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithWriter(os.Stderr),
		stdouttrace.WithoutTimestamps())
	if err != nil {
		log.Fatalf("creating stdoutmetric exporter: %v", err, stdoutTrace)
	}

	// TODO: detect the env variables or use dynamic config

	// Span processor for tracez - this is an interface that gets called on each span.
	zpagesSpanProcessor := zpages.NewSpanProcessor()
	// HTTP handler that dumps the data in the span processor
	thandler := zpages.NewTracezHandler(zpagesSpanProcessor)
	http.DefaultServeMux.Handle("/debug/tracez", thandler)

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(r),
		trace.WithBatcher(stdoutTrace),
		trace.WithSpanProcessor(zpagesSpanProcessor),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{}))

	exTracer := otel.Tracer("start")
	_, span := exTracer.Start(ctx, "startSpan")
	span.End()

	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}
}

// Should be called only if another main function didn't already initialize this.
func InitOTelHost() {
	// Host telemetry -
	host.Start()

	if err := runtime.Start(
		runtime.WithMinimumReadMemStatsInterval(time.Second),
	); err != nil {
		log.Fatalln("failed to start runtime instrumentation:", err)
	}
}

// OTel is the struct containing the bootstrapped telemetry
// This is called from main() to setup the SDKs and providers in a particular way.
type OTel struct {
	Registry *prom.Registry
}

func InitOTel(ctx context.Context, serviceName string) (func(), error) {
	ot := &OTel{}
	r := resource.NewWithAttributes(semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName))

	tracingCleanup := InitTracing(ctx, r)

	InitOTelHost()

	// =========== Metrics
	// Default to prometheus, which is the most common and may be used in other libraries.
	// Collector can handle pushing to other destinations.

	// prom.DefaultRegisterer
	ot.Registry = prom.NewRegistry()

	// Register otel metrics with prometheus, via a collector
	promExporter, err := prometheus.New(prometheus.WithRegisterer(ot.Registry))
	if err != nil {
		return nil, err
	}

	// Set the global meter provider - most libraries should use it directly.
	// metric SDK provides an implementation with all the bells, including sending to prom.
	provider := metric.NewMeterProvider(metric.WithReader(promExporter))
	global.SetMeterProvider(provider)

	ot.InitProm()

	// Init an OTel metric
	m1 := global.MeterProvider().Meter("test1")
	c1, _ := m1.Int64Counter("counter1")
	c1.Add(ctx, 1)

	// End telemetry magic
	return func() {
		provider.Shutdown(ctx)
		tracingCleanup()
	}, nil
}

func (ot *OTel) InitProm() {
	reg := ot.Registry
	// Add Go module build info.
	reg.MustRegister(collectors.NewBuildInfoCollector())
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	//Latencies := prometheus.NewHistogramVec(prometheus.HistogramOpts{
	//	Namespace: ns,
	//	Subsystem: "h2",
	//	Name:      "http_request_lat",
	//	Help:      "",
	//}, []string{})
	//
	//reg.MustRegister(Latencies)

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))

}
