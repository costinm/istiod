# Basic install

Minimal install for in-cluster, running or debugging istiod outside of the cluster.

```shell

kubectl apply -k github.com/kubernetes-sigs/gateway-api/config/crd?rev=v0.3.0
kubectl apply -k github.com/kubernetes-sigs/gateway-api/config/crd

# Takes forever (clone repo): kubectl apply -k github.com/istio/istio/manifests/charts/base

# Includes 'legacy' service accounts, permissions - and 'ignore' validation webhook
kubectl apply -f https://raw.githubusercontent.com/istio/istio/master/manifests/charts/base/files/gen-istio-cluster.yaml

```

# Istiod useful options

By default Istiod with no settings will run in a config close to the 'default' install.

When running on a VM (or local debugging), few extra options can be useful:


```shell

# Disable config patching - should be done by in-cluster or installer
# This is also useful when running with lower permissions.
export VALIDATION_WEBHOOK_CONFIG_NAME=
export INJECTION_WEBHOOK_CONFIG_NAME=

 # Skip authentication for XDS requests, for debugging without tokens
 export XDS_AUTH=false
 
 # Authenticate, even for plaintext 
 export XDS_AUTH_PLAINTEXT=true
```


Istiod uses the default ports 
- 9876 (controlz), 
- 15010/15012 (XDS, XDS-TLS), 
- 8080 (debug),
- 15017 (https/inject),
- 15014 (monitoring/prom)


For debug, it may load a mesh config from ./etc/istio/config/mesh - replacing any in-system config.

# Server identification

control_plane:{identifier:"{\"Component\":\"istiod\",\"ID\":\"\",\"Info\":{\"version\":\"unknown\",\"revision\":\"unknown\",\"golang_version\":\"go1.16\",\"status\":\"unknown\",\"tag\":\"unknown\"}}

POD_NAME is used for ID
istio.io/pkg/version is populated from ldflags.

# Agent options

```shell

# No mTLS for control plane
export USE_TOKEN_FOR_CSR=true
export USE_TOKEN_FOR_XDS=true

```
