#FROM istionightly/ci:2019-08-14 AS istio-build
#RUN make

#RUN GOPATH="" go get -v -u istio.io/istio/pilot/cmd/pilot-agent

####

FROM golang:1.13-alpine AS build
RUN apk add --no-cache git
#WORKDIR /src
#COPY go.mod go.sum ./
#RUN go mod download
COPY . .

# Runs in /go directory
RUN GOPROXY=https://proxy.golang.org GO111MODULE=on CGO_ENABLED=0 GOOS=linux \
  GOPATH="" go build -a -ldflags '-extldflags "-static"' ./cmd/hyperistio

RUN GOPROXY=https://proxy.golang.org GO111MODULE=on CGO_ENABLED=0 GOOS=linux \
  GOPATH="" go build -a -ldflags '-extldflags "-static"' ./cmd/istiok8s

#### Fortio - used for the tests
FROM fortio/fortio:latest AS fortio

# Istio proxy using upstream envoy/alpine
FROM envoyproxy/envoy-alpine AS proxy-alpine

### Container with SSH for testing from VM (exposing ports)
FROM envoyproxy/envoy-alpine AS envoy-gw
RUN apk add --no-cache git openssh

### Container for testing for Istio VM deb files, in a bionic container
FROM ubuntu:bionic AS sidecar-test

RUN apt-get update &&\
    apt-get install -y curl iptables iproute2 &&\
    curl  https://storage.googleapis.com/istio-release/releases/1.3.0-rc.0/deb/istio-sidecar.deb -o /tmp/istio.deb &&\
    dpkg -i /tmp/istio.deb

COPY --from=fortio /usr/share/fortio /usr/share/fortio
COPY --from=fortio /usr/bin/fortio /usr/bin/fortio

EXPOSE 8079
EXPOSE 8080
EXPOSE 8081

### Container running the combined control plane
FROM envoyproxy/envoy-alpine AS istio-control

COPY --from=build /go/hyperistio /usr/local/bin/hyperistio
COPY --from=build /go/istiok8s /usr/local/bin/istiok8s
ENTRYPOINT /usr/local/bin/istiok8s
