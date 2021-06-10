# Default based on ubuntu:bionic
ARG BASE=istio/proxyv2:1.10.0

FROM gcr.io/dmeshgate/ugate:latest as ugate

# Alternative
#RUN  curl -L https://storage.googleapis.com/istio-release/releases/1.10.0/deb/istio-sidecar.deb  -o /tmp/istio.deb &&\
#    dpkg -i /tmp/istio.deb && rm /tmp/istio.deb

# Alternative
#FROM istio/proxyv2:latest AS istio
#COPY --from=istio /usr/local/bin/envoy /usr/local/bin/
#COPY --from=istio /usr/local/bin/envoy /usr/local/bin/
#...


FROM ${BASE}

# Many are already included in proxyv2:
RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt install -y --no-install-recommends \
      less \
      openssh-server \
      \
      ca-certificates \
      curl \
      iptables \
      iproute2 \
      iputils-ping \
      knot-dnsutils \
      netcat \
      tcpdump \
      conntrack \
      bsdmainutils \
      net-tools \
      lsof \
      linux-tools-generic \
      sudo \
      && update-ca-certificates \
      && apt-get upgrade -y \
      && apt-get clean \
      && rm -rf  /var/log/*log /var/lib/apt/lists/* /var/log/apt/* /var/lib/dpkg/*-old /var/cache/debconf/*-old

RUN useradd -u 1000 -U -d /work -s /bin/bash app && \
     usermod -G users app && \
      echo "app ALL=NOPASSWD: ALL" >> /etc/sudoers

# Already created by useradd
#RUN useradd -m --uid 1337 istio-proxy && \
#  echo "istio-proxy ALL=NOPASSWD: ALL" >> /etc/sudoers


RUN mkdir -p /etc/certs && \
    mkdir -p /run/ssh && \
    mkdir -p /etc/istio/proxy && \
    mkdir -p /etc/istio/config && \
    mkdir -p /var/lib/istio/envoy && \
    mkdir -p /var/lib/istio/config && \
    mkdir -p /var/lib/istio/proxy && \
    chown -R 1337 /etc/certs /etc/istio /var/lib/istio && \
    mkdir /run/ugate && \
    mkdir /run/sshd && \
    chmod 700 /run/sshd

WORKDIR /

EXPOSE 8080

# Include SSH for debugging.
COPY sshd_config /etc/ssh/sshd_config
COPY run.sh /usr/local/bin/run.sh


# Test only
RUN echo 'root:test' | chpasswd

# Not sure if this is still needed and why...
RUN sed 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' -i /etc/pam.d/sshd &&\
    mkdir -p  /var/run/secrets/kubernetes.io/serviceaccount

# Replace bootstrap
COPY envoy_bootstrap.json /var/lib/istio/envoy/envoy_bootstrap_tmpl.json
COPY --from=ugate /ko-app/ugate /usr/local/bin/ugate
ENV CA_ROOT_CA=SYSTEM \
    XDS_ROOT_CA=SYSTEM

ENTRYPOINT ["/usr/local/bin/run.sh"]
