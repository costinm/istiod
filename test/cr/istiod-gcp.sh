#!/bin/bash

env

if [[ -n ${PROJECT} ]]; then
  echo gcloud container clusters get-credentials ${CLUSTER} --zone ${ZONE} --project ${PROJECT}
  gcloud container clusters get-credentials ${CLUSTER} --zone ${ZONE} --project ${PROJECT}
fi


# Disable webhook config patching - manual configs used, proper DNS certs means no cert patching.
export VALIDATION_WEBHOOK_CONFIG_NAME=
export INJECTION_WEBHOOK_CONFIG_NAME=

# TODO: get secrets as well, maybe file from storage
export DISABLE_LEADER_ELECTION=true

# Disable the DNS-over-TLS server
export DNS_ADDR=

if [[ -n ${MESH} ]]; then
  echo ${MESH} > /etc/istio/config/mesh
else
  cat /etc/istio/config/mesh_template.yaml | envsubst > /etc/istio/config/mesh
fi

echo Starting with: $*

exec /usr/local/bin/pilot-discovery discovery \
   --httpsAddr OFF \
   --secureGRPCAddr OFF \
   ${EXTRA_ARGS} $*
