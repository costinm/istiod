#!/usr/bin/env bash

export SKIP_CLEANUP=1
export INSTANCE_IP=${IP:-10.52.3.116}
export POD_NAMESPACE=none
export ISTIO_META_INTERCEPTION_MODE=NONE
export ISTIO_META_HTTP10=1
export ISTIO_META_CONFIG_NAMESPACE=none

$TOP/out/linux_amd64/release/pilot-agent proxy --discoveryAddress localhost:16010  --proxyLogLevel info

# ISTIO_CLUSTER_CONFIG=remote.cluster.env EXEC_USER=costin bash -x /usr/local/bin/istio-start.sh run
