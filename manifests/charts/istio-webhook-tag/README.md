# Istio Channel

This chart installs MutatingWebhooks for Istio injection, providing the helm
equivalent of 'istioctl x revision tag set' command described in
[Flexible Revisions Labels/Revision Tags](https://docs.google.com/document/d/13IGuJg8swtLdNGW5cpF7ZdVkgge8voNp9DWBD93Wb1Q/edit#heading=h.xw1gqgyqs5b).


Example for creating a tag 'stable', pointing to the default istiod install:

```shell

helm -n istio-system upgrade --install istio-webhook-stable manifets/charts/istio-webhook-tag \
  --set tag=stable

```

Creating a tag 'canary', pointing to the v1-10 istiod install:

```shell

helm -n istio-system upgrade --install istiod-v1-10 ../istio/manifests/charts/istio-control/istio-discovery \
  --set revision=v1-10
 
helm -n istio-system upgrade --install istio-webhook-canary manifests/charts/istio-webhook-tag \
  --set tag=canary --set revision=v1-10

```

After the 'canary' is tested, users can point 'stable' to the new revision:

```shell

helm -n istio-system upgrade --install istio-webhook-stable manifets/charts/istio-webhook-tag \
  --set tag=stable --set revision=v1-10

```


This chart can also be used as a basic for customizing the labeling and selection. For example,
this chart adds an additional namespace key, "istio.io/tag", which behaves in the same way with 
istio.io/rev, but allows workloads to override. 

For backward compatibility reasons, if the 'istio.io/rev' key is used on a namespace, it will
take precendence over Pods with a different 'istio.io/rev'


# Backward-compatibility mode

This chart also supports configuring the original Istio injection
labeling scheme, using 'istio-injection' namespace label and 
'sidecar.istio.io/inject' workload label.

If installing the chart with the name "default", it will configure
a mutating webhook overriding the 'istio-sidecar-injector', and 
supporting:

- the namespace label 'istio-injection=enabled'
- object label: 'sidecar.istio.io/inject=true'

Before installing the chart in default mode, if you have an existing
Istio installed with Istio operator or helm you need to migrate the
object:

```shell

kubectl  annotate mutatingwebhookconfiguration istio-sidecar-injector app.kubernetes.io/managed-by=Helm
kubectl  annotate mutatingwebhookconfiguration istio-sidecar-injector meta.helm.sh/release-name=istio-webhook-default
kubectl  annotate mutatingwebhookconfiguration istio-sidecar-injector meta.helm.sh/release-namespace=istio-system
kubectl  label mutatingwebhookconfiguration istio-sidecar-injector app.kubernetes.io/managed-by=Helm


helm upgrade --install istio-webhook-default manifests/charts/istio-webhook-tag \
  --set enableIstioInjection=true --set revision=v1-10

```

It is also possible to provide 'install all workloads by default' mode, where any workload
or namespace not explicitly labeled will be injected:

```shell

helm upgrade --install istio-webhook-default manifests/charts/istio-webhook-tag \
  --set enableNamespacesByDefault=true --set revision=v1-10

```
