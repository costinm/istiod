#!/bin/bash

env

if [[ -n ${PROJECT} ]]; then
  echo gcloud container clusters get-credentials ${CLUSTER} --zone ${ZONE} --project ${PROJECT}
  gcloud container clusters get-credentials ${CLUSTER} --zone ${ZONE} --project ${PROJECT}
fi

export ISTIOD_PORT=${ISTIOD_PORT:-443}

# Disable webhook config patching - manual configs used, proper DNS certs means no cert patching.
export VALIDATION_WEBHOOK_CONFIG_NAME=
export INJECTION_WEBHOOK_CONFIG_NAME=

# TODO: get secrets as well, maybe file from storage
export DISABLE_LEADER_ELECTION=true

export USE_TOKEN_FOR_CSR=true
export USE_TOKEN_FOR_XDS=true

# Disable the DNS-over-TLS server
export DNS_ADDR=

if [[ -n ${MESH} ]]; then
  echo ${MESH} > /etc/istio/config/mesh
else
  cat /etc/istio/config/mesh_template.yaml | envsubst > /etc/istio/config/mesh
  cat /etc/istio/config/mesh
fi

kubectl get ns istio-system
if [[ "$?" == "1" ]]; then
  kubectl create ns istio-system
  kubectl apply -k github.com/istio/istio/manifests/charts/base
fi

# Make sure the mutating webhook is installed, and prepare CRDs
# This also 'warms' up the kubeconfig - otherwise gcloud will slow down startup of istiod.
kubectl get mutatingwebhookconfiguration istiod-managed
if [[ "$?" == "1" ]]; then
  echo "Mutating webhook missing, initializing"
  # TODO: include the charts in the image !
  cat /var/lib/istio/inject/mutating_template.yaml | envsubst > /var/lib/istio/inject/mutating.yaml
  cat /var/lib/istio/inject/mutating.yaml
  kubectl apply -f /var/lib/istio/inject/mutating.yaml
else
  echo "Mutating webhook found"
fi
echo Starting with: $*


# TODO: if injection template, injection values are present in the cluster, get them and use instead of the
# built-in templates. Same for mesh config.

exec /usr/local/bin/pilot-discovery discovery \
   --httpsAddr OFF \
   --secureGRPCAddr OFF \
   ${EXTRA_ARGS} ${LOG_ARGS} $*
