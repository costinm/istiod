# How to enable core dump for istiod-injected sidecars

1. Edit the injection template, and add:

TODO: symlink from /var/lib/istio/core.proxy to the /etc/istio/proxy dir for compat with tools looking there.


```yaml

template:
    initContainers:
      - name: enable-core-dump
        args:
        - -c
        - sysctl -w kernel.core_pattern=/etc/istio/proxy && ulimit -c unlimited
        command:
          - /bin/sh
        image: istio/proxy_v2:1.3.0
        resources: {}
        securityContext:
          runAsUser: 0
          runAsNonRoot: false

```

2. To use tproxy, add

```yaml


template:
    containers:
    - name: istio-proxy
      securityContext:
        capabilities:
          add:
          - NET_ADMIN
      env:
      - name: ISTIO_META_INTERCEPTION_MODE
        value: TPROXY

```

3. To override bootstrap, create a configmap with 2 files, envoy_bootstrap_tmpl.json and envoy_bootstrap_drain.json
The name of the map should be specified in the pod using "sidecar.istio.io/bootstrapOverride" annotation.
