# istiod

Implementation for [Isto-SDS](https://docs.google.com/document/d/1X4QNWSr0aoT2eK-f5a6ZgWgX8VXP-suQbfO-SjBozyw/edit#)
and [Simplified Istio/istiod](https://docs.google.com/document/d/1v8BxI07u-mby5f5rCruwF7odSXgb9G8-C9W5hQtSIAg/edit#)

The repository contains `istiod` - a component linking multiple Istio micro-services in a single binary, with
settings optimized for production use and operations. 

There are 2 variants - one intended to be run on K8S, loading configs and endpoints from one or more apiservers and 
exposing them via MCP and XDS. It includes certificate generation and SDS, using K8S-signed certificates for bootstrap.
The second variant - `istiod-vm` - is intended to be run on VMs and non-K8S environments, and does not link or make calls to apiservers, instead gets all config and endpoints from files or MCP sources. 

Both components can be run as standalone binaries on a VM or in a docker container, as well as deployed in a K8S cluster.
The K8S cluster running `istiod` can optionally be separate from the K8S clusters running workloads.

*istiod does not use CLI parameters*, and has defaults reflecting 'best practices' for production use, with minimal use
of alpha features. All configurations are loaded from mesh.yaml and component-specific config files. The binary will also
use files from filesystem and the common environment variables - like $HOME/.kube/config and $KUBECONFIG, /etc/certs, etc. 

istiod will not handle TLS - but will start an envoy sidecar, similar with current Istio. This allows using SDS 
for certificates and additional load blancing and envoy features.


# WIP: Configuration and environment

The binary can run in 'root' mode, with $CWD set as "/", or as non-root, with the $CWD set to the HOME of the user
running istio. For consistency, it will use the same filesystem layout in both - for example certs will be in /etc/certs in
root mode, and $PWD/etc/certs in non-root mode. 

The combined binary does not use CLI - only config files:

- ./etc/istio/mesh.yaml - mesh config, with support for reload. Includes additional component-specific yaml files.
- ./etc/istio/conf/ - CRDs loaded by Galley (for file-based binary). TODO: will be combined with k8s
- ./etc/certs/ - certificates using a custom CA ( citadel, etc )

Since Istio and Apiserver codebases are used, some environment variables used in Istio will be respected, but 
the intent is to have all runtime behavior described by mesh.yaml and independent of how the binary is started.

Ideally user should only run `istiod` with no CLI/env, and have a working control plane with the 
recommended/best practice config. 

# Galley integration

Galley Processing2 is loaded at startup, and initializes the filesystem and mesh config Sources. In `istiod` binary the
in-cluster K8S is attempted first, falling back to $KUBECONFIG and $HOME/.kube. 

The 'istiod-vm' variant will not attempt to connect to k8s and will not link the client. It will still support K8S APIs
that are used in Istio - EndpointSet, Ingress, etc. The ./etc/istio/conf will be loaded automatically, with remaining 
config and endpoints using MCP, using either `istiod` or alternate implementations. 

Pilot will initially load using localhost grpc, with the plan to move to 'direct' calls to Galley Source to avoid extra overhead and scale better. An additional registry and config provider for Pilot will be created, backed by Galley Source,
with the plan to deprecate/remove the old Pilot K8S sources.



