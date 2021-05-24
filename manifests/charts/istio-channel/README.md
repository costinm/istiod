# Istio Channel

This chart installs MutatingWebhooks for Istio injection using 
stable 'channel' labels. This is the helm implementation equivalent of 
the 'istioctl tag' command described in [Flexible Revisions Labels/Revision Tags](https://docs.google.com/document/d/13IGuJg8swtLdNGW5cpF7ZdVkgge8voNp9DWBD93Wb1Q/edit#heading=h.xw1gqgyqs5b)
proposal.


Creating a tag 'stable', pointing to the default istiod install:

```shell

helm -n istio-system upgrade --install istio-webhook-stable manifets/charts/istio-channel --set channel=stable 

```

Creating a tag 'canary', pointing to the v1.11-dev istiod install:

```shell

helm -n istio-system upgrade --install istiod-v1-11-dev ../istio/manifests/charts/istio-control/istio-discovery --set revision=v1-11-dev 
 
helm -n istio-system upgrade --install istio-webhook-canary manifests/charts/istio-channel --set channel=canary --set revision=v1-11-dev

```

For each channel, you should use namespace or workload label "istio.io/rev=CHANNEL" to activate injection
for the entire namespace or specific pods. The pod label has priority over namespace, so in a namespace
with label 'istio.io/rev=stable', a pod with the label 'istio.io/rev=canary' will use the canary version.


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



helm upgrade --install default manifests/charts/istio-channel \
  --version v1-10

```

