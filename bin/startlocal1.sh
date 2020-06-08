
mkdir -p out/certs/vm
generate_cert --mode citadel \
	    -out-cert out/certs/vm/cert-chain.pem -out-priv out/certs/vm/key.pem \
	        -host spiffe://cluster.local/ns/vmtest/sa/default

kubectl get cm -n istio-system istio-ca-root-cert -ojsonpath='{.data.root-cert\.pem}' > out/certs/vm/root-cert.pem

OUTPUT_CERTS=./out/certs/vm PROV_CERT=./out/certs/vm PROXY_CONFIG="$(cat ~/kube/local/proxyconfig.yaml | envsubst)" go run ./pilot/cmd/pilot-agent proxy sidecar --templateFile ./tools/packaging/common/envoy_bootstrap_v2.json


