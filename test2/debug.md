I=/workspace/istio-master/go/src/github.com/costinm/istio-install


./bin/iop istio-gateway gateway ${BASE}/gateways/istio-ingress --set global.proxy.accessLogFile=/dev/stdout -f user-values-ingress.yaml
iop istio-control istio-discovery istio-control/istio-discovery --set global.proxy.accessLogFile=/dev/stdout


helm template --name istio --namespace istio-system . -f $I/test/istio-system-values.yaml > $I/test/istio-system-1.1.yaml

kubectl apply --prune -l release=istio -f  $I/test/istio-system-1.1.yaml 
