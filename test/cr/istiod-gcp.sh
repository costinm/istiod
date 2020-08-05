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

export REVISION=${REV:-managed}
export GKE_CLUSTER_URL=https://container.googleapis.com/v1/projects/${PROJECT}/locations/${ZONE}/clusters/${CLUSTER}
export TRUST_DOMAIN=${PROJECT}.svc.id.goog
export CA_PROVIDER=${CA_PROVIDER:-istiod}
export CA_ADDR=${CA_ADDR:-}

if [[ "${ASM}" == "1" ]]; then
  export STACKDRIVER=1
  export CA_ADDR=meshca.googleapis.com:443
fi

# TODO:
# - copy inject template and mesh config to cluster (first time) or from cluster
# - revision support
# - option to enable 'default' ingress class
# - telemetry on stackdriver
# - option for managed CA
# - support non-GKE clusters - fetch a kubeconfig from secret manager
# - set trustDomain to match managed CA by default ( no need for cluster.local )

kubectl get ns istio-system
if [[ "$?" != "0" ]]; then
  echo "Initializing istio-system and CRDs, fresh cluster"
  kubectl create ns istio-system
  #kubectl apply -k github.com/istio/istio/manifests/charts/base
  kubectl apply -f /var/lib/istio/config/gen-istio-cluster.yaml \
      --record=false --overwrite=false   --force-conflicts=true --server-side
fi

if [[ -n ${MESH} ]]; then
  echo ${MESH} > /etc/istio/config/mesh
else
  cat /etc/istio/config/mesh_template.yaml | envsubst > /etc/istio/config/mesh
  cat /etc/istio/config/mesh
fi

cat /var/lib/istio/inject/values_template.yaml | envsubst > /var/lib/istio/inject/values

# TODO: istio must watch it - no file reloading
kubectl get -n istio-system cm istio-${REVISION}
if [[ "$?" != "0" ]]; then
  echo "Initializing revision"
  kubectl -n istio-system create cm istio-${REVISION} --from-file /etc/istio/config/mesh

  # Sidecars will report to stackdriver - requires proper setup.
  if [[ "${STACKDRIVER}" == "1" ]]; then
    cat /var/lib/istio/config/telemetry-sd.yaml | envsubst | kubectl apply -f -
  else
    # Prometheus only.
    cat /var/lib/istio/config/telemetry.yaml | envsubst | kubectl apply -f -
  fi
fi


# Make sure the mutating webhook is installed, and prepare CRDs
# This also 'warms' up the kubeconfig - otherwise gcloud will slow down startup of istiod.
kubectl get mutatingwebhookconfiguration istiod-${REVISION}
if [[ "$?" == "1" ]]; then
  echo "Mutating webhook missing, initializing"
  # TODO: include the charts in the image !
  cat /var/lib/istio/inject/mutating_template.yaml | envsubst > /var/lib/istio/inject/mutating.yaml
  cat /var/lib/istio/inject/mutating.yaml
  kubectl apply -f /var/lib/istio/inject/mutating.yaml
else
  echo "Mutating webhook found"
fi
echo Starting $*

# What audience to expect for Citadel and XDS - currently using the non-standard format
# TODO: use https://... - and separate token for stackdriver/managedCA
export TOKEN_AUDIENCE=${TRUST_DOMAIN}

# Istiod will report to stackdriver
export ENABLE_STACKDRIVER_MONITORING=${ENABLE_STACKDRIVER_MONITORING:-1}

# TODO: if injection template, injection values are present in the cluster, get them and use instead of the
# built-in templates. Same for mesh config.

exec /usr/local/bin/pilot-discovery discovery \
   --httpsAddr OFF \
   --trust-domain ${TRUST_DOMAIN} \
   --secureGRPCAddr OFF \
   --grpcAddr OFF \
   ${EXTRA_ARGS} ${LOG_ARGS} $*
