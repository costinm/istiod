###############################################################################
#### Fortio - used for the tests
FROM fortio/fortio:latest AS fortio

###############################################################################
### Container for testing for Istio VM deb files, in ubuntu container
## Will include hyperistio (the minimal, non-k8s version), fortio for testing, other tools needed for debugging.
## Will install envoy from the official release .deb ( but not using the systemd unit for start )
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
