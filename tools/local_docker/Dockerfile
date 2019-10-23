FROM alpine

WORKDIR /
RUN mkdir -p /etc/certs && \
    mkdir -p /etc/istio/proxy && \
    mkdir -p /etc/istio/config && \
    mkdir -p /var/lib/istio/envoy && \
    mkdir -p /var/lib/istio/config && \
    mkdir -p /var/lib/istio/proxy && \
    chown -R 1337 /etc/certs /etc/istio /var/lib/istio

COPY ./istiod /usr/local/bin/istiod
COPY ./var/lib/istio /var/lib/istio/

USER 1337:1337
ENTRYPOINT /usr/local/bin/istiod
