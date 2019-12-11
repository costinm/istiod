# istiod

Implementation for [Isto-SDS](https://docs.google.com/document/d/1X4QNWSr0aoT2eK-f5a6ZgWgX8VXP-suQbfO-SjBozyw/edit#)
and [Simplified Istio/istiod](https://docs.google.com/document/d/1v8BxI07u-mby5f5rCruwF7odSXgb9G8-C9W5hQtSIAg/edit#)

# Setup 

## Basic install for istiod

This is the 'default' install, creating an istiod and ingress deployment in istio-system.

The install can be done in a fresh cluster, or in a cluster where istio is already setup - the install is not
interfering with the normal istio install. 

1. Cluster-wide settings - require cluster admin, grant broad permissions. This step 
needs to be repeated on each release, all instances of the control plane will use the same CRDs.

```bash

kubectl apply -k github.com/costinm/istiod/kustomize/cluster

# Customize the mutating webhook to select which workloads/namespaces will be selected.
# Default is namespaces with istio-env=istiod label.

kubectl apply -k github.com/costinm/istiod/kustomize/autoinject

```

2. Install istiod 

```bash

kubectl apply -k github.com/costinm/istiod/kustomize/istiod

```

3. Install an ingress gateway 


```bash

kubectl apply -k github.com/costinm/istiod/kustomize/isto-ingress

```


## Testing environment 

This installs istiod, knative, 2 namespaces running fortio servers and client - one with secure and one insecure.
More tests and scenarios will be added. This is intended to be used in the 'stability/perf/scale' clusters. 

Note: These steps must be run after Istiod is in a 'Running' state. Istiod patches the mutatingwebhook resource to add CA credentials. Without those credentials, Kubernetes will refuse to create pods that run through the webhook. If you installed the workloads too early, you may need to delete stuck replicasests in order for them to start trying to create pods again. 

1. Cluster-wide settings - requires cluster-admin

```bash

kubectl apply -k github.com/costinm/istiod/test/all-cluster

```

2. Everything else

```bash

kubectl apply -k github.com/costinm/istiod/test/all

```


# Missing features 

- Galley validation not yet integrated

- SDS code change to read from a file if secure JWT are not available WIP


