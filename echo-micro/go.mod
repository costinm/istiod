module github.com/costinm/grpc-mesh/echo-micro

go 1.16

replace github.com/costinm/grpc-mesh/bootstrap => ../bootstrap

replace github.com/costinm/grpc-mesh/telemetry/otelgrpc => ../telemetry/otelgrpc

replace github.com/costinm/grpc-mesh/telemetry/logs/zap => ../telemetry/logs/zap

replace github.com/costinm/grpc-mesh/gen/proto => ../gen/proto

replace github.com/costinm/grpc-mesh/gen/grpc-go => ../gen/grpc-go

replace github.com/costinm/grpc-mesh => ../

//replace google.golang.org/grpc => ../../grpc

//replace github.com/GoogleCloudPlatform/cloud-run-mesh => ../../cloud-run-mesh

//replace github.com/GoogleCloudPlatform/cloud-run-mesh/pkg/k8s => ../../cloud-run-mesh/pkg/k8s

require (
	contrib.go.opencensus.io/exporter/prometheus v0.4.0
	github.com/costinm/grpc-mesh/bootstrap v0.0.0-00010101000000-000000000000
	github.com/costinm/grpc-mesh/telemetry/logs/zap v0.0.0-00010101000000-000000000000
	github.com/costinm/grpc-mesh/telemetry/otelgrpc v0.0.0-00010101000000-000000000000
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/prometheus/client_golang v1.12.1
	go.opencensus.io v0.23.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.28.0
	go.opentelemetry.io/contrib/instrumentation/host v0.31.0
	go.opentelemetry.io/contrib/zpages v0.31.0
	go.opentelemetry.io/otel v1.6.1
	go.opentelemetry.io/otel/exporters/prometheus v0.28.0
	go.opentelemetry.io/otel/metric v0.28.0
	go.opentelemetry.io/otel/sdk v1.6.1
	go.opentelemetry.io/otel/sdk/metric v0.28.0
	go.uber.org/zap v1.21.0
	golang.org/x/sync v0.1.0
	google.golang.org/grpc v1.52.0
)

require (
	github.com/GoogleCloudPlatform/cloud-run-mesh v0.0.0-20220128230121-cac57262761b
	github.com/costinm/grpc-mesh/gen/grpc-go v0.0.0-00010101000000-000000000000
	github.com/costinm/grpc-mesh/gen/proto v0.0.0-00010101000000-000000000000
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/zapr v1.2.3
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	k8s.io/klog/v2 v2.60.1
)
