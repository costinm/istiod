# Tests for the 'safe' profile of Istiod

The chart will install a number of workloads and configurations
to verify interoperability with ambient and with gRPC using
Istio and TD control planes. 

"Safe" Istio is a config mode where the behavior of sidecars
is modified to match ambient and GKE - for example cluster.local
will be local by default, and a number of features that only
work with Istio sidecar are disabled (including EnvoyFilters).

# Testing gRPC

One of the priorities is validating the stability of 
gRPC with GAMMA, which currently has little coverage. 

This is WIP, needs a client driver to generate client 
traffic and metrics. 

Example commands to test gRPC manually:

```shell

ns=istio-safe-stability
POD=$(kubectl --namespace=$ns get -l app=echo,version=v1 pod -o=jsonpath='{.items[0].metadata.name}')

kubectl  -n $ns exec -it ${POD}  -- client --qps 2 --msg hi  grpc://echo-grpc-v1:7070

Or using grpcurl: 

grpcurl -plaintext  -d '{"url": "xds:echo-grpc-v2.istio-safe-stability.svc.cluster.local:7070"}' echo-grpc-v2:7070 proto.EchoTestService/ForwardEcho 
 
grpcurl -vv -plaintext  -d '{"url": "grpc://echo-grpc-v1:7070"}' echo-grpc-v2:7070 proto.EchoTestService/ForwardEcho
 
 ```
