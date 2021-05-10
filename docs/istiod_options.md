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
 
 # 
```

# Agent options

```shell

# No mTLS for control plane
export USE_TOKEN_FOR_CSR=true
export USE_TOKEN_FOR_XDS=true

```
