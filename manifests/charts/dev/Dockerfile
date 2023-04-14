# Docker image based on Istio build tools, with extra tooling.

FROM gcr.io/istio-testing/build-tools:master-latest-amd64

# Extra tools
RUN apt update && apt install -y --no-install-recommends \
    less net-tools sudo lsof netcat tcpdump iptables dnsutils iproute2 vim dnsutils \
    sshd syncthing && \
    apt-get clean && \
         rm -rf \
        	/tmp/* \
        	/var/lib/apt/lists/* \
        	/var/tmp/*

# Primary user
RUN useradd -u 1000 -U -d /work -s /bin/bash build && \
     usermod -G users build && \
      echo "build ALL=NOPASSWD: ALL" >> /etc/sudoers

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOPROXY=https://proxy.golang.org
ENV HOME=/work
WORKDIR /work
