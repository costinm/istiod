#!/bin/sh

# Prepare istio for 'standalone' ( outside of k8s ) run.
# This can be used on a dev VM or in a docker container.

# It is possible to convert this to go - but as a shell it
# is easier to understand the steps and tune.


# If an existing Istio is running, grab the root CA.
# It should be rotated and migrated to intermediary CA
# Stored as istio-ca-secret, not mounted (watched).
# Files: ca-cert.pem, root-cert.pem, ca-key.pem
function getOldRootCA() {

}

function generateIntermediaryCA() {

}

# Get the intermediary CA from the cluster.
#
# Stored in 'cacerts' volume, mounted as /etc/cacerts
function getIntermediaryCA() {

}

# Get the long-lived JWT token from istiod secret.
# This is mounted as ./var/run/secrets/...
# Looks like:
# {
#  "alg": "RS256",
#  "kid": "Du6ZWVsySmrW6gxW5Z1CIS4smFAaVxrbZqPOet57iVo"
# }.
# {
#  "iss": "kubernetes/serviceaccount",
#  "kubernetes.io/serviceaccount/namespace": "istio-system",
#  "kubernetes.io/serviceaccount/secret.name": "istiod-service-account-token-nr5br",
#  "kubernetes.io/serviceaccount/service-account.name": "istiod-service-account",
#  "kubernetes.io/serviceaccount/service-account.uid": "4e37f80a-2cf2-4d5c-a037-252df8f7bebe",
#  "sub": "system:serviceaccount:istio-system:istiod-service-account"
# }
function getIstiodK8SToken() {
  SECRET_NAME=$(kubectl -n istio-system get sa istiod-service-account -ojsonpath='{.secrets[0].name}')
  TOKEN=$(kubectl -n istio-system get secret ${SECRET_NAME}  -o jsonpath="{.data['token']}" | base64 --decode)
  # TODO: error handling.
  mkdir -p ./var/run/secrets/kubernetes.io/serviceaccount/
  echo -n $TOKEN > ./var/run/secrets/kubernetes.io/serviceaccount/token
  return $TOKEN
}

# Get the user-configured inject map.
# Stored as 'istio-sidecar-injector-REVISION'
# If none, a default will be provisioned.
function getInjectMap() {

}

# Get the initial mesh config.
# TODO: this should be watched by Istiod
function getMeshConfig() {

}
