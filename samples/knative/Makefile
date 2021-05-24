CA_ADDR ?= istiod-xds-icq63pqnqq-uc.a.run.app:443
SERVICE_ACCOUNT ?= default
INTERMEDIATE_KEYSZ ?= 4096
INTERMEDIATE_ORG ?= Istio
INTERMEDIATE_CN ?= Intermediate CA

# Get a token
%/token:
	echo '{"kind":"TokenRequest","apiVersion":"authentication.k8s.io/v1","spec":{"audiences":["istio-ca"], "expirationSeconds":2592000}}' | \
    		kubectl create --raw /api/v1/namespaces/$(dir $@)serviceaccounts/default/token -f - | \
    		jq -j '.status.token' > $(dir $@)istio-token
	kubectl -n istio-system get secret istio-ca-secret -ojsonpath='{.data.ca-cert\.pem}' | \
      	 base64 -d > root-cert.pem

# Get a cert from the CA, using the token. Alternative to downloading the root CA and generating the cert locally.
%/token-cert: L=$(dir $@)
%/token-cert: %/workload.csr %/token  %/key.pem
	echo -n '{"csr": "' >$(dir $@)/csr-escaped
	cat  $< | awk -F'\\n' '{printf "%s\\n",$$0} END {print "\"}"}'  >> $(dir $@)/csr-escaped
	cat $(dir $@)csr-escaped | grpcurl  -import-path . \
        -proto ./istioca.proto -use-reflection=false \
         -d @ -H "Authorization:  Bearer $(shell cat $(dir $@)istio-token )" \
         ${CA_ADDR} istio.v1.auth.IstioCertificateService/CreateCertificate > $(dir $@)/cert.json
	cat default/cert.json |jq .certChain[0] | sed 's/\\n/\n/g' | sed 's/"//' > $(dir $@)/workload-cert.pem
	cat default/cert.json |jq .certChain[1] | sed 's/\\n/\n/g' | sed 's/"//' > $(dir $@)/root-cert.pem

%/workload.csr: L=$(dir $@)
%/workload.csr: %/key.pem %/workload.conf
	@echo "generating $@"
	@openssl req -new -config $(L)/workload.conf -key $< -out $@

%/key.pem:
	@echo "generating $@"
	@mkdir -p $(dir $@)
	@openssl genrsa -out $@ 4096

%/workload.conf: L=$(dir $@)
%/workload.conf:
	@echo "[ req ]" > $@
	@echo "encrypt_key = no" >> $@
	@echo "prompt = no" >> $@
	@echo "utf8 = yes" >> $@
	@echo "default_md = sha256" >> $@
	@echo "default_bits = $(INTERMEDIATE_KEYSZ)" >> $@
	@echo "req_extensions = req_ext" >> $@
	@echo "x509_extensions = req_ext" >> $@
	@echo "distinguished_name = req_dn" >> $@
	@echo "[ req_ext ]" >> $@
	@echo "subjectKeyIdentifier = hash" >> $@
	@echo "basicConstraints = critical, CA:false" >> $@
	@echo "keyUsage = digitalSignature, keyEncipherment" >> $@
	@echo "extendedKeyUsage = serverAuth, clientAuth" >> $@
	@echo "subjectAltName=@san" >> $@
	@echo "[ san ]" >> $@
	@echo "URI.1 = spiffe://cluster.local/ns/$(L)sa/$(SERVICE_ACCOUNT)" >> $@
	@echo "DNS.1 = $(SERVICE_ACCOUNT).$(L:/=).svc.cluster.local" >> $@
	@echo "[ req_dn ]" >> $@
	@echo "O = $(INTERMEDIATE_ORG)" >> $@
	@echo "CN = $(INTERMEDIATE_CN)" >> $@
	@echo "L = $(L:/=)" >> $@
