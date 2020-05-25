
setup:
	kubectl get secret istio-ca-secret -n istio-system -o "jsonpath={.data['ca-cert\.pem']}" | base64 -d > tls.crt
	kubectl get secret istio-ca-secret -n istio-system -o "jsonpath={.data['ca-key\.pem']}" | base64 -d > tls.key
	kubectl -n istio-system create secret generic istio-certmanager-ca --from-file tls.crt=tls.crt --from-file ca.crt=tls.crt --from-file tls.key=tls.key
	rm tls.key tls.crt

cm-install: setup
	kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v0.15.0/cert-manager.yaml
	kubectl -n istio-system apply -f istio-issuer.yaml
