# Default based on ubuntu:bionic
ARG BASE=gcr.io/istio-testing/proxyv2:latest

# Alternative
#RUN  curl -L https://storage.googleapis.com/istio-release/releases/1.10.0/deb/istio-sidecar.deb  -o /tmp/istio.deb &&\
#    dpkg -i /tmp/istio.deb && rm /tmp/istio.deb

# Alternative
#FROM istio/proxyv2:latest AS istio
#COPY --from=istio /usr/local/bin/envoy /usr/local/bin/
#COPY --from=istio /usr/local/bin/envoy /usr/local/bin/
#...


FROM ${BASE}

#RUN useradd -u 1000 -U -d /work -s /bin/bash app && \
#     usermod -G users app && \
#      echo "app ALL=NOPASSWD: ALL" >> /etc/sudoers

# Replace bootstrap
COPY envoy_bootstrap.json /var/lib/istio/envoy/envoy_bootstrap_tmpl.json

#RUN mkdir -p /etc/certs && \
#    mkdir -p /run/ssh && \
#    mkdir -p /etc/istio/proxy && \
#    mkdir -p /etc/istio/pod && \
#    mkdir -p /etc/istio/config && \
#    mkdir -p /var/lib/istio/envoy && \
#    mkdir -p /var/lib/istio/config && \
#    mkdir -p /var/lib/istio/proxy && \
#    chown -R 1337 /etc/certs /etc/istio /var/lib/istio && \
#    mkdir /run/ugate && \
#    mkdir /run/sshd && \
#    chmod 700 /run/sshd
#
#RUN usermod -a -G tty istio-proxy

ENTRYPOINT ["/usr/local/bin/run.sh"]
