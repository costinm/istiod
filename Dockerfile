# Moved the base image - including mod download - to separate Dockerfile.
###############################################################################
FROM costinm/istiod-build:latest AS build

COPY cmd ./cmd
COPY pkg ./pkg

# Runs in /go directory
RUN GOPROXY=https://proxy.golang.org CGO_ENABLED=0 GOOS=linux \
  GOPATH="" go build -a -ldflags '-extldflags "-static"' -o istiod-vm ./cmd/istiod-vm
RUN GOPROXY=https://proxy.golang.org CGO_ENABLED=0 GOOS=linux \
  GOPATH="" go build -a -ldflags '-extldflags "-static"' -o istiod ./cmd/istiod && ls


###############################################################################
FROM envoyproxy/envoy AS envoy
###############################################################################
FROM gcr.io/distroless/cc:latest as distroless

COPY --from=build /ws/istiod /usr/local/bin/istiod
COPY --from=envoy /usr/local/bin/envoy /usr/local/bin/envoy

WORKDIR /

COPY ./var/lib/istio/envoy/* /var/lib/istio/envoy/

USER 1337:1337
ENTRYPOINT /usr/local/bin/istiod


###############################################################################
### Container running the combined control plane, with an alpine base ( smaller than distroless but with shell )
### TODO: add a distroless variant.
### This image should work as a drop-in replacement for Pilot, Galley(MCP portion), WebhookInjector
### Citadel, Gallye/Validation remain as separate deployments.
FROM envoyproxy/envoy-alpine AS istio-control

COPY --from=build /ws/istiod /usr/local/bin/istiod

WORKDIR /
RUN mkdir -p /etc/certs && mkdir -p /etc/istio/proxy && mkdir -p /etc/istio/config && mkdir -p /var/lib/istio/envoy && \
    chown -R 1337 /etc/certs /etc/istio /var/lib/istio

# Defaults
COPY ./etc/istio/config/mesh /etc/istio/config/mesh
COPY ./var/lib/istio/envoy/* /var/lib/istio/envoy/
USER 1337:1337
ENTRYPOINT /usr/local/bin/istiod
