###############################################################################
FROM costinm/istiod-build:latest AS build

COPY cmd ./cmd
COPY pkg ./pkg

# Runs in /go directory
RUN go build -a -ldflags '-extldflags "-static"' -o istiod-vm ./cmd/istiod-vm
RUN go build -a -ldflags '-extldflags "-static"' -o istiod ./cmd/istiod && ls


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

