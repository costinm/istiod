Will install versioned variants of the echo app.

- golang
- cpp
- TODO: java

Versions:
  - with istio sidecar
  - proxyless, with istio sidecar
  - proxyless, no agent
  - proxyless, TD
  - raw

# CPP

Istio echo, grpc cpp version. No CLI, only listens gRPC on 7070

Injected with istio agent for bootstrap generation, secrets, XDS proxy

Based on https://github.com/grpc/grpc/blob/master/test/cpp/interop/istio_echo_server.cc
Should mirror istio test 'app'
