# Istio Echo app

This is a copy of istio/istio/pkg/test/echo, the main Istio test appliction, extracted
as a standalone module and exported for use in tests or as a standalone app with minimal deps. 

Istio contains a lot of useful code, but it is currently organised as a single huge module
with far too many dependencies. Until this changes, it is useful to extract useful subsets.

## Changes from original

- exported
- commented out http3 forwarder in server/forwarder/http3, the tls hack doesn't work well with 1.17
- remove deps


## Dependency removal

Example: 
```shell
$ go mod why github.com/Azure/go-autorest

# github.com/Azure/go-autorest
github.com/costinm/grpc-mesh/echo/client
istio.io/istio/pkg/test/framework/components/cluster
istio.io/istio/pkg/kube
k8s.io/client-go/plugin/pkg/client/auth
k8s.io/client-go/plugin/pkg/client/auth/azure
github.com/Azure/go-autorest/autorest
github.com/Azure/go-autorest
```

This shows echo depends on kube API, which brings a lot of dependencies, via 'cluster' -
commented out the unused methods.

go mod why github.com/Azure/go-autorest
# github.com/Azure/go-autorest
github.com/costinm/grpc-mesh/echo/server/endpoint
istio.io/istio/pkg/istio-agent/grpcxds
istio.io/istio/pilot/pkg/model
istio.io/istio/pilot/pkg/model.test
istio.io/istio/pilot/pkg/xds
istio.io/istio/pkg/kube
k8s.io/client-go/plugin/pkg/client/auth
k8s.io/client-go/plugin/pkg/client/auth/azure
github.com/Azure/go-autorest/autorest
github.com/Azure/go-autorest


