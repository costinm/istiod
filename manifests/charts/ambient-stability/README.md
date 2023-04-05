# Tests for ztunnel

- to install on GKE - Calico network and eBPF works
- metrics can be used to see if interception works
- istioctl logs works


- https://github.com/istio/ztunnel/blob/eb55427c4c21631ac81343098b06210537c013d3/src/admin.rs


```shell
istioctl pc log ztunnel-x6ml2.istio-system --level debug

istioctl dashboard prometheus
istioctl dashboard kiali

```

# Configs 

- threads

## Calico

- needs cni.projectcalico allocation and eBPF

