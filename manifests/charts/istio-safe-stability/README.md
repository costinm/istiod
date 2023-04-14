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


# Testing gRPC

```shell

ns=istio-safe-stability
POD=$(kubectl --namespace=$ns get -l app=echo,version=v1 pod -o=jsonpath='{.items[0].metadata.name}')
kubectl  -n $ns exec -it ${POD}  -- client --qps 2 --msg hi  grpc://echo-grpc-v1:7070


grpcurl -plaintext  -d '{"url": "xds:echo-grpc-v2.istio-safe-stability.svc.cluster.local:7070"}' echo-grpc-v2:7070 proto.EchoTestService/ForwardEcho 
 
grpcurl -vv -plaintext  -d '{"url": "grpc://echo-grpc-v1:7070"}' echo-grpc-v2:7070 proto.EchoTestService/ForwardEcho
 
 ```
