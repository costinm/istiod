module github.com/costinm/grpc-mesh/echo

go 1.19

//replace github.com/costinm/grpc-mesh/bootstrap => ../bootstrap

require (
	contrib.go.opencensus.io/exporter/prometheus v0.4.0
	//github.com/costinm/grpc-mesh/bootstrap v0.0.0-00010101000000-000000000000
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/go-multierror v1.1.1
	github.com/prometheus/client_golang v1.12.1
	github.com/spf13/cobra v1.3.0
	go.opencensus.io v0.23.0
	golang.org/x/net v0.7.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/grpc v1.46.2
	google.golang.org/protobuf v1.27.1
	istio.io/istio v0.0.0-20220128192255-56edd90bb930
	istio.io/pkg v0.0.0-20220127152359-bf0ff5d5e4ca
	k8s.io/utils v0.0.0-20220127004650-9b3446523e65
)

require github.com/costinm/grpc-mesh/bootstrap v0.0.0-20230218062247-c645305d763c

require (
	cloud.google.com/go v0.100.2 // indirect
	cloud.google.com/go/compute v1.0.0 // indirect
	cloud.google.com/go/logging v1.4.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.2 // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cncf/udpa/go v0.0.0-20210930031921-04548b0d99d4 // indirect
	github.com/cncf/xds/go v0.0.0-20220112060520-0fa49ea1db0c // indirect
	github.com/costinm/meshauth v0.0.0-20230123031534-9e635566c01e // indirect
	github.com/envoyproxy/go-control-plane v0.10.2-0.20220325020618-49ff273808a1 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.6.2 // indirect
	github.com/go-kit/log v0.1.0 // indirect
	github.com/go-logfmt/logfmt v0.5.0 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/googleapis/gax-go/v2 v2.1.1 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/prometheus/statsd_exporter v0.21.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.1 // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/api v0.65.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220118154757-00ab72f36ad5 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/klog/v2 v2.40.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
