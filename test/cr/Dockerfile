FROM google/cloud-sdk:slim

RUN mkdir -p /var/run/secrets && \
    mkdir -p /var/run/istio/inject && \
    mkdir -p /var/run/istio/config && \
    mkdir -p /etc/istio
RUN apt-get update && apt-get install -y gettext-base
RUN apt-get update && \
    apt-get install --no-install-recommends -y \
      gettext-base kubectl \
   && apt-get clean \
   && rm -rf  /var/log/*log /var/lib/apt/lists/* /var/log/apt/* /var/lib/dpkg/*-old /var/cache/debconf/*-old


COPY mesh_template.yaml /etc/istio/config/mesh_template.yaml
COPY istiod /usr/local/bin/pilot-discovery

COPY istiod-gcp.sh /usr/local/bin/istiod-gcp.sh
COPY injection-template.yaml /var/lib/istio/inject/config
COPY injection-values.yaml /var/lib/istio/inject/values
COPY mutating_template.yaml /var/lib/istio/inject/mutating_template.yaml

ENTRYPOINT /usr/local/bin/istiod-gcp.sh
