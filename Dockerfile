FROM istionightly/ci:2019-08-14 AS istio-build
#RUN make

#RUN GOPATH="" go get -v -u istio.io/istio/pilot/cmd/pilot-agent

####

FROM golang:1.13-alpine AS build
RUN apk add --no-cache git
#WORKDIR /src
#COPY go.mod go.sum ./
#RUN go mod download
#COPY . .
#RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' \
#        -o /bin/a.out ./cmd/cloudshell_open

RUN GOPATH="" go get -v -u istio.io/istio/pilot/cmd/pilot-agent


####

#### Fortio

FROM fortio/fortio:latest AS fortio

# Istio proxy using upstream envoy/alpine
FROM envoyproxy/envoy-alpine AS proxy-alpine


### SSH for testing
FROM envoyproxy/envoy-alpine AS envoy-gw
RUN apk add --no-cache git openssh


###


### Test for Istio sidecar

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
