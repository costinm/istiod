#!/bin/bash

if [ "$1" == "istiod" ]; then

  make pilot-discovery

  export ISTIOD_ADDR=istiod.istio-system.svc:15012
  export WEBHOOK=istiod
  cd /

  /work/out/linux_amd64/pilot-discovery discovery -n istio-system

elif [ "$1" == "init" ]; then

  ln -s /work/out/linux_amd64/envoy /usr/local/bin
  mkdir -p /var/lib/istio/envoy
  ln -s /work/tools/packaging/common/envoy_bootstrap_v2.json /var/lib/istio/envoy/envoy_bootstrap_tmpl.json

elif [ "$1" == "install" ]; then

  ln -s /work/out/linux_amd64/envoy /usr/local/bin
  mkdir -p /var/lib/istio/envoy
  ln -s /work/tools/packaging/common/envoy_bootstrap_v2.json /var/lib/istio/envoy/envoy_bootstrap_tmpl.json

  apt-get update && apt-get install -y --no-install-recommends \
    dnsutils vim net-tools iptables

elif [ "$1" == "dig" ]; then

  dig @localhost -p 15053 www.google.com
  dig @localhost -p 15053 istiod.istio-system.svc.cluster.local

elif [ "$1" == "istio-agent" ]; then
  make pilot-agent

  cd /

  /work/out/linux_amd64/pilot-agent proxy

fi
