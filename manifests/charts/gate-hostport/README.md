# Minimalistic Istio Gateway running as hostport

Primary use case: expose a docker registry on the node, on 
port 5000.

It runs in a separate namespace, with no special permissions.

```shell

helm upgrade -n istio-gw-hostport --install istio-gw-hostport .

```

This can also be used to expose other ports on the node,
if the node has public IP addresses. Note that firewalls
may need to be adjusted.

