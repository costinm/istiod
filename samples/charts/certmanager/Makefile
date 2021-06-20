

# This will convert the istio self-signed root CA into a CertManager
# secret.
convert-root:
	kubectl get secret istio-ca-secret -n istio-system -o "jsonpath={.data['ca-cert\.pem']}" | base64 -d > tls.crt
	kubectl get secret istio-ca-secret -n istio-system -o "jsonpath={.data['ca-key\.pem']}" | base64 -d > tls.key
	kubectl -n istio-system create secret generic istio-certmanager-ca --from-file tls.crt=tls.crt --from-file ca.crt=tls.crt --from-file tls.key=tls.key || true
	rm tls.key tls.crt

# Install cert manager. Older: 0.15.0
cm-install: 
	kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v1.3.1/cert-manager.yaml
	kubectl -n istio-system apply -f istio-issuer.yaml

# Follow KNative setup
dns-setup:
	kubectl -n cert-manager create secret generic clouddns-dns01-solver-svc-acct  \
  		--from-file=key.json
