# Running in-cluster local registry

Will start a docker registry in the given namespace plus a port forwarder.

Docker allows localhost to be used as 'insecure' - it relies on the host having the port open.
A daemon set is used to forward 'localhost' requests to the actual registry.



```shell

 kubectl apply -k github.com/costinm/istiod/test/registry

 POD=$(kubectl get pods --namespace istio-system -l app=kube-registry \
            -o template --template '{{range .items}}{{.metadata.name}} {{.status.phase}}{{"\n"}}{{end}}' \
            | grep Running | head -1 | cut -f1 -d' ')

 kubectl port-forward --namespace istio-system $POD 5000:5000 &

```

# TODO

- run a Gateway (envoy) as port forwarder, with full Istio config and mTLS to the registry
- configure the forwarder to allow MTLS/TLS/auth, using istio API.
- configure the proxy to forward to per-namespace registries ? 

The idea would be to have a volume in each namespace with the binary images and .tar.gz 
