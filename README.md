# istio-vm

Run full Istio control plane on a VM, outside K8S. It can also run in a pod in a remote (secure) K8S cluster, separating
the control plane from the managed workload clusters.

This is an experimental project - it is also experimenting with using upstream envoy configs by default, so no support 
for mixer plugin or authn-jwt.  




