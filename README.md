# istio-vm

Run full Istio control plane on a VM, outside K8S. It can also run in a pod in a remote (secure) K8S cluster, separating
the control plane from the managed workload clusters.

The new control plane consists of a unified binary, with simplified configuration. To variants are provided, both
including all micro-services required (galley, pilot, certs, sds):

1. File-based, using Galley file-based configs and remote MCP sources. Certificates will need to be 
provisioned by external CAs and stored in files before startup to enable TLS.

2. K8S-based, using KUBECONFIG or incluster config and remote MCP or K8S registry sources. This includes
code to auto-generate a certificate signed by K8S.

In both cases all control plane services will share the same gRPC port (15010, 15011), using Envoy to handle
TLS if configured. 

This is an experimental project - it is also experimenting with using upstream envoy configs by default, so no support 
for mixer plugin or authn-jwt.  

# Configuration

The combined binary does not use CLI - only config files:

- ./etc/istio/mesh.yaml - mesh config, with support for reload. 
- ./etc/istio/conf/ - CRDs loaded by Galley (for file-based binary). TODO: will be combined with k8s
- ./etc/certs/ - certificates using a custom CA ( citadel, etc )

Running the binary with CWD=/ will use the system-wide configs.

Additional yaml config files for the Pilot, Galley options will be read while transition to mesh.yaml is in progress.

Since Istio and Apiserver codebases are used, some environment variables used in Istio will be respected, but 
the intent is to have all runtime behavior described by mesh.yaml and independent of how the binary is started.

Ideally user should only run "./hyperistio" with no CLI/env, and have a working control plane with the 
recommended/best practice config. 


# Galley integration

Galley Processing2 is loaded at startup, and initializes the filesystem and mesh config Sources. If K8S is enabled,
it'll also load the K8S config source, and add it to GalleyServer.Sources.
