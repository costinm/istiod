#!/bin/bash

# Startup for the dev image
# This includes tools to help debug and even source code to make changes.
# It includes a sshd server, to allow access even without kube.


cd /
env
echo Starting $*


#export PROXY_CONFIG="{"discoveryAddress": "istiod-costin-asm1-big1-xds-icq63pqnqq-uc.a.run.app:443","proxyBootstrapTemplatePath": "./tools/packaging/common/envoy_bootstrap.json" 'binary_path':'./out/linux_amd64/release/envoy'}"
export PROXY_CONFIG='{"discoveryAddress": "'${ISTIOD}'"}'

mkdir -p /var/run/secrets/kubernetes.io/serviceaccount
echo $JWT > /var/run/secrets/kubernetes.io/serviceaccount/token

# TODO: auto-detect
# User system certificates for CA connection
export CA_ROOT_CA=SYSTEM
export XDS_ROOT_CA=SYSTEM
# The JWT is long lived - until we implement metadata and OIDC
export JWT_POLICY=first-party-jwt
# TODO: should be default
export PILOT_CERT_PROVIDER=istiod

# TODO: app=fortio, pod name derived using some unique id
export POD_NAME=${POD_NAME:-unset}
export POD_NAMESPACE=${POD_NAMESPACE:-httpbin}
export TRUST_DOMAIN=${TRUST_DOMAIN:-cluster.local}

# Capture DNS
export ISTIO_META_DNS_CAPTURE=true

# Get bootstrap from pilot, can be patched with EnvoyFilter
export BOOTSTRAP_XDS_AGENT=true

function start_agent() {
  cat > /etc/istio/pod/labels << EOF
app="$POD_NAME"
security.istio.io/tlsMode="istio"
version=v1
EOF
  #pilot-agent istio-iptables
  pilot-agent proxy sidecar  --domain ${POD_NAMESPACE}.svc.cluster.local &
  # TODO: wait for ready before returning
   #--proxyLogLevel=info --proxyComponentLogLevel=misc:info --log_output_level=default:debug
}

function start_gw() {
  cat > /etc/istio/pod/labels << EOF
istio="ingressgateway"
version=v1
EOF
  pilot-agent proxy router --domain ${POD_NAMESPACE}.svc.cluster.local
   #--proxyLogLevel=info --proxyComponentLogLevel=misc:info --log_output_level=default:debug
}

function start_ssh() {
  # /etc dir may be RO, use var/run

  # TODO: custom call to get a cert for SSH from istiod
  # File will exist in k8s if a secret is mounted
  if ! test -f /run/ssh/ssh_host_ecdsa_key; then
      ssh-keygen -q -f /run/ssh/ssh_host_ecdsa_key -N '' -t ecdsa
  fi

  # TODO: support certificates for client auth
  if ! test -f /run/ssh/authorized_keys; then
      echo ${SSH_AUTH} >> /run/ssh/authorized_keys
  fi

  # Set correct right to ssh keys
  chown -R root:root /run/ssh /run/sshd
  chmod 0700 /run/ssh
  chmod 0600 /run/ssh/*

  chmod 700 /run/sshd

  echo "======== Starting SSHD with ${SSH_AUTH}"

  /usr/sbin/sshd
}

start_ssh


# Debug entrypoint (reverse H2)
export BASE_PORT=14000
/usr/local/bin/ugate &

if [[ -z "$*" ]] ; then
  start_gw
else
  start_agent
  $*
fi
