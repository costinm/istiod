Simplified istio gateway deployment, to be used as-is or as a base for your own charts.

Will only install only the Istio proxy workloads and associated Autoscale and PodDisruptionBudget. 

It is strongly recommended to not install gateways in istio-system - this chart allows it
for very special cases where it is not possible to migrate the IP address.

A Service and Gateway configs with the appropriate selector needs to be
created by user.  The installed deployment can be selected using "istio: NAMESPACE" label.

This deployment uses SDS by default - no Secret or Volume mounts, and is optimized for 
simplicity and to be used as a base to be extended by a user's own charts.

# Options

Common options:

- revision - select the istiod revision to use, by setting the injection labels on the workload. By default it
  uses the 'sidecar.istio.io/inject=true' label.

- serviceAccountName - if set, will use an existing service account. User can grant RBAC permissions and create 
  the service account independently.
  
- imagePullSecrets - if set and if serviceAccountName is not specified, will create a dedicated service account
  for the workloads, with the ImagePullSecrets set.

# Changes compared with istio/manifests/charts/istio-ingress:

- "customService" is set, the Service should be deployed as a separate
install. It is recommended to maintain the Service and Gateway objects in the
same chart and independent of the actual workload, so the ports are in sync. 

- flattened and simplified options - the chart is intended for one
gateway deployment and nothing else.

- name of the install (Release.Name) is used as name for all generated
  objects, there is no overlap with other installs.

- only injected deployment supported.

- removed role and automatic creation of a service account.
  Instead, using SDS for getting secrets. A 'serviceAccount' option
  added, defaulting to "default" service account in the namespace.

# Migration from old charts

The new chart has different naming and does not interfere with the old charts - it is 
installed in a separate namespace and has simplified config.

User will install the new chart in separate namespaces, create the Service/Gateway configs,
test and switch the DNS to the new Service. 


# istio-system migration

This chart can also be installed in istio-system. This is not recommended, since
the control plane and secrets are highly sensitive and should be kept separated.

If preserving the IP address that was allocated is a requirement, you can install
in istio-system. 

When installing in istio-system, the name of the gateway and labels will match
the existing one, i.e. "istio: ingressgateway"

# Replacing istio-ingressgateway deployment

To replace the existing default deployment, use:

```
kubectl -n istio-system annotate poddisruptionbudget istio-ingressgateway app.kubernetes.io/managed-by=Helm
kubectl -n istio-system annotate poddisruptionbudget istio-ingressgateway meta.helm.sh/release-name=istio-ingressgateway
kubectl -n istio-system annotate poddisruptionbudget istio-ingressgateway meta.helm.sh/release-namespace=istio-system
kubectl -n istio-system label poddisruptionbudget istio-ingressgateway app.kubernetes.io/managed-by=Helm

kubectl -n istio-system annotate hpa istio-ingressgateway app.kubernetes.io/managed-by=Helm
kubectl -n istio-system annotate hpa istio-ingressgateway meta.helm.sh/release-name=istio-ingressgateway
kubectl -n istio-system annotate hpa istio-ingressgateway meta.helm.sh/release-namespace=istio-system
kubectl -n istio-system label hpa istio-ingressgateway app.kubernetes.io/managed-by=Helm

kubectl -n istio-system annotate deploy istio-ingressgateway app.kubernetes.io/managed-by=Helm
kubectl -n istio-system annotate deploy istio-ingressgateway meta.helm.sh/release-name=istio-ingressgateway
kubectl -n istio-system annotate deploy istio-ingressgateway meta.helm.sh/release-namespace=istio-system
kubectl -n istio-system label deploy istio-ingressgateway app.kubernetes.io/managed-by=Helm

helm -n istio-system upgrade --install istio-ingressgateway manifests/charts/gateway-workload

```
