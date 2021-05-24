
# Istio Channel

This chart installs MutatingWebhooks for Istio injection using 
stable 'channel' labels.

The chart is opinionated, but you can use it as an example to define
any other labeling scheme.


# Backward-compatibility mode

This chart also supports configuring the original Istio injection
labeling scheme.

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

