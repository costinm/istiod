# Examples for using persistent sessions

This is focused on using the sessions in a Gateway and Waypoint.

Sidecars can be used, but it also works well with ambient and zTunnel and
with CNI-provided encryption.

## Headers vs Cookies

Cookies would work automatically in browsers. Unless they are blocked, or there are 
requirements for user consent and UI. Cookie jars are complicated too. The Path 
can't be effectively be used with multiple services using the same cookie name.

Headers are more flexible, and are more aligned with the rest of Istio design - the
'session' header is in fact just metadata about the client IP selected. With 
ambient PTR-DS we can convert this back to identity of the peer, it is metadata
we already exchange for telemetry.

It was added in https://github.com/envoyproxy/envoy/pull/23145, Oct 11 2022 for 1.24, and 
is supported by Istio 1.16. Note that while Istio 1.15 or lower is in cluster in any revision,
the validator in 1.15 will reject the config since it's not yet supported.

## Header names

This example uses the convention:

- VERSION--SERVICE to represent subsets
- "x-istio-cluster" for the cluster name - if we add support for setting just the version we can switch to 
'knative-serving-tag' or an equivalent including just the version.

## Reusing telemetry

Istio telemetry sets a few headers:
- 

## Debug examples

```shell

# Install CRDs
kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd/experimental" | kubectl apply -f -

# Verify cluster is set as a header
curl http://localhost:9880/debug -v 2>&1 |egrep "x-istio|set-cookie"

# Find the gateway pod
POD=$(kubectl -n session get pod -l istio=session -o jsonpath='{.items[0].metadata.name}')

# Confirm filter is added on listener
istioctl pc l ${POD}.session --port 8080 -o yaml |grep stateful_session -A 10

# Show routes
istioctl pc r ${POD}.session --name http.8080 -o yaml

curl http://localhost:9880/debug -v -H "x-istio-cluster: outbound|8080||v2--fortio.session.svc.cluster.local" -H "Cookie: session-fortioroute2=MTAuNDguMC4xNzE6ODA4MA==" 
```
