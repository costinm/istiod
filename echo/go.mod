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
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/grpc v1.46.2
	google.golang.org/protobuf v1.27.1
	istio.io/istio v0.0.0-20220128192255-56edd90bb930
	istio.io/pkg v0.0.0-20220127152359-bf0ff5d5e4ca
	k8s.io/utils v0.0.0-20220127004650-9b3446523e65
)

require github.com/costinm/grpc-mesh/bootstrap v0.0.0-20230218062247-c645305d763c

require github.com/costinm/meshauth v0.0.0-20230123031534-9e635566c01e // indirect
