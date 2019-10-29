# istiod

Implementation for [Isto-SDS](https://docs.google.com/document/d/1X4QNWSr0aoT2eK-f5a6ZgWgX8VXP-suQbfO-SjBozyw/edit#)
and [Simplified Istio/istiod](https://docs.google.com/document/d/1v8BxI07u-mby5f5rCruwF7odSXgb9G8-C9W5hQtSIAg/edit#)

# Setup and gaps

istio/istio now has most of the code, but configs are not yet merged.

1. Cluster-wide settings

```bash

kubectl apply -k github.com/costinm/istiod/kustomize/cluster

```

2. Install istiod (plain and easy)

```bash
kubectl apply -k github.com/costinm/istiod/kustomize/istiod
```

Alternative - and example for how to kustomize the install:

```bash

kubectl apply -k github.com/costinm/istiod/test/k8s/istiod

```

The files in test/k8s/istiod contain a custom version of the injection template and mesh config.
We still use values.yaml - only for injection, until it is cleaned up.

3. Autoinjection

This step is not yet automated - you will need the file from ./kustomize/cluster-autoinject/mutatingwebhook.yaml
or ./kustomize/autoinject/mutatingwebhook.yaml. First file enables cluster-wide autoinject, the second requires
labels on the namespaces.

The current manual step is to extract the K8S cert - for example from you .kube/config or $KUBECONFIG file.

```yaml
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS...
```  

Copy the string under clusters.cluster[N].certificate-authority-data, and paste it under 
webhooks.clientConfig.caBundle


This step is currently moving to istioctl - no plan to use auto-enabling of the webhook (for now).

# SDS 

The pilot-agent and injection template used in this repo are SDS-only, using an SDS server started 
in pilot-agent, using a local UDS socket ( /etc/istio/proxy/SDS ). 

Istiod includes a subset of Citadel - to generate a self-signed root CA and provide it to pilot-agent,
using the K8S JWT token to identify the workloads. 

The workloads connect to istiod using K8S-signed certificates.


