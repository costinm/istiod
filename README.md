# istiod

Implementation for [Isto-SDS](https://docs.google.com/document/d/1X4QNWSr0aoT2eK-f5a6ZgWgX8VXP-suQbfO-SjBozyw/edit#)
and [Simplified Istio/istiod](https://docs.google.com/document/d/1v8BxI07u-mby5f5rCruwF7odSXgb9G8-C9W5hQtSIAg/edit#)

# Setup 

## Basic install for istiod

This is the 'default' install, creating an istiod and ingress deployment in istio-system.

The install can be done in a fresh cluster, or in a cluster where istio is already setup - the install is not
interfering with the normal istio install. 

1. Cluster-wide settings

```bash

kubectl apply -k github.com/costinm/istiod/kustomize/cluster

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

1. Cluster-wide settings - requires cluster-admin

```bash

kubectl apply -k github.com/costinm/istiod/test/all-cluster

```

2. Everything else

1. Cluster-wide settings - requires cluster-admin

```bash

kubectl apply -k github.com/costinm/istiod/test/all

```


# Gaps 

- Galley validation not yet integrated

- SDS code change to read from a file if secure JWT are not available WIP

