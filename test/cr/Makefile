ISTIO_REGION ?= us-central1
ISTIO_PROJECT ?= wlhe-cr

ZONE ?= us-central1-c
#PROJECT ?= costin-asm1
#CLUSTER ?= big1
PROJECT_ID ?= costin-asm1
CLUSTER ?= big1

CPU ?= 1
MEM ?= 2G

# Works with telemetry/stackdriver
#REV ?= m4
#TAG ?= cr

# WIP: stackdriver+managedca
REV ?= asm
TAG ?= asm

HUB ?= gcr.io/${ISTIO_PROJECT}
export TAG
export HUB
export REV
# Suffix for the domain URLs -
DOMAIN?=-icq63pqnqq-uc.a.run.app
#SDOMAIN=-t2mbbtjbiq-uc.a.cloud-run-qual.sandbox.google.com
URL=istiod-${PROJECT_ID}-${CLUSTER}-${REV}${DOMAIN}

# Default to the docker env dir
ISTIO_SRC ?= /work

# Main testing target for ASM
all: build/cloudrun run_asm test/run

proxy:
	cd ${ISTIO_SRC} && HUB=${HUB} TAG=${TAG} make push.docker.proxyv2
	docker tag ${HUB}/proxyv2:${TAG} costinm/proxyv2:${TAG}
	docker push costinm/proxyv2:${TAG}
	# Kick the pod to force reload
	kubectl -n fortio-${REV} delete po --all --force --grace-period=0

gcp/init:
	gcloud services enable staging-run.sandbox.googleapis.com
	gcloud config set run/region us-central1
	gcloud config set run/platform managed
	gcloud config set project costin-asm1
	#https://gce-staging.pantheon-ci.corp.google.com/run?project=costin-asm1
	#gcloud config set api_endpoint_overrides/run https://staging-run.sandbox.googleapis.com/

cr/status:
	gcloud run services --platform managed list

# Command - to override the default command ( pilot )

# --concurrency= = 0 unspec, set a max per instance
# --service-account - create one per tenant
# --timeout - 10m5s

#{"node": {"id": "sidecar~1.1.1.1~debug~cluster.local", "metadata": {"GENERATOR": "event"}},"typeUrl": "istio.io/connections"}

TOP=$(shell pwd)

build/telemetry-gen:
	(cd ${ISTIO_SRC} && helm3 template -n istio-system istio-default \
		manifests/charts/istio-control/istio-discovery \
		--set telemetry.v2.stackdriver.enabled=true --set telemetry.v2.prometheus.enabled=true \
        --set telemetry.v2.prometheus.wasmEnabled=true --set telemetry.v2.accessLogPolicy.enabled=true \
        --set revision=XREVISION --set telemetry.v2.metadataExchange.wasmEnabled=true  \
        --set global.multiCluster.clusterName=XCLUSTER ) | sed s/XREVISION/\$${REVISION}/ > istio-all.yaml

# Build Istiod only - this is what goes into the docker image.
build/istiod:
	(cd ${ISTIO_SRC}; go build -o ${TOP}/istiod ./pilot/cmd/pilot-discovery)


#export DOCKER_HOST=tcp://localhost:2375

build/cloudrun: build/istiod
	docker build -t ${HUB}/pilot:${TAG} .
	docker push  ${HUB}/pilot:${TAG}

# Run 2 revisions - one ASM one OSS
# The OSS revision can be used in non-GKE clusters
# The ASM will use ManagedCA and Stackdriver, on GKE
run:
	#gcloud config unset api_endpoint_overrides/run
	REV=oss $(MAKE) _run
	REV=asm ENVEXTRA="ASM=1," CPU=4 MEM=4G $(MAKE) _run

# Run just the ASM version.
run_asm:
	#gcloud config unset api_endpoint_overrides/run
	ENVEXTRA="ASM=1," $(MAKE) _run

test/run:
	REV=asm $(MAKE) fortio
	REV=oss $(MAKE) fortio

# Run fortio in fortio-$REV, with $REV istio env
fortio:
	cat test/fortio.yaml | envsubst | kubectl apply -f -

# Common target for running
_run:
	#gcloud config unset api_endpoint_overrides/run
	# Slow start - but pinned
	# Alternative: --cpu 4 --memory 4G
	gcloud alpha run deploy istiod-${PROJECT_ID}-${CLUSTER}-${REV} --allow-unauthenticated \
	 --project ${ISTIO_PROJECT} --region ${ISTIO_REGION} --platform managed \
     --image gcr.io/${ISTIO_PROJECT}/pilot:${TAG} \
     --set-env-vars=${ENVEXTRA}REV=${REV},TAG=${TAG},CLUSTER=${CLUSTER},ZONE=${ZONE},PROJECT=${PROJECT_ID},ISTIOD_DOMAIN=${DOMAIN} \
	 --port 8080 --concurrency 1000 --timeout 900 --cpu ${CPU} --memory ${MEM} --max-instances=1

#--cpu 4 --memory 4G --min-instances=1

# Max set to 1 to test single-instance perf

	#gcloud config set api_endpoint_overrides/run https://staging-run.sandbox.googleapis.com/
#	gcloud alpha run deploy istiod${REV}-${PROJECT_ID}-${CLUSTER}-xds --allow-unauthenticated \
#	 --project ${ISTIO_PROJECT} --region ${ISTIO_REGION} --platform managed \
#        --image gcr.io/${ISTIO_PROJECT}/pilot:${TAG} \
#       --set-env-vars=CLUSTER=${CLUSTER},ZONE=${ZONE},PROJECT=${PROJECT_ID},ISTIOD_DOMAIN=${DOMAIN},INJECT_ENABLED=false,VALIDATION_ENABLED=false \
#        --port 15010 --cpu 4 --memory 4G --min-instances=1q
	#gcloud config unset api_endpoint_overrides/run

build/staging:
	PROJECT_ID=costin-asm1 make build/remote


staging: build/staging run_staging

push: build/remote run run_staging

logs:
	 gcloud logging read 'resource.type = "project" OR resource.type = "cloud_run_revision"'

cert:
	echo {} | grpcurl  -import-path /work/security/proto/ -proto /work/security/proto/istioca.proto -use-reflection=false -v -d @ -H "Authorization:  Bearer $(gcloud auth print-identity-token)" -rpc-header "authorization: Bearer $(gcloud auth print-identity-token)"  istiodxds-b3lo6zbfmq-uc.a.run.app:443 istio.v1.auth.IstioCertificateService/CreateCertificate


# Verify injection is active
test/inject:
	time curl -v https://istiod${REV}-${PROJECT_ID}-${CLUSTER}${DOMAIN}/inject  -d @mutating_req.json -HContent-Type:application/json | jq .response.patch | sed s/\"// | base64 -d | jq .


test/dump:
	curl -H "Authorization: Bearer $(gcloud auth print-identity-token)" https://${URL}:443/debug/configz

test/ep:
	curl -H "Authorization: Bearer $(gcloud auth print-identity-token)" https://${URL}:443/debug/connectionz

test/grpc:
	 echo {} | grpcurl -v -d @ ${URL}:443 envoy.service.discovery.v3.AggregatedDiscoveryService/StreamAggregatedResources


# Inject on a local workload
local/inject:
	curl  http://localhost:8080/inject/  -d @mutating_req.json -HContent-Type:application/json  | jq .response.patch | sed s/\"// | base64 -d | jq .

local/grpc:


# Steps:
# - create cluste
# - grant IAM to cloud run to cluster
# - base install
GOOGLE_APPLICATION_CREDENTIALS=${HOME}/.config/gcloud/legacy_credentials/costin@google.com/adc.json
local:
	#docker stop istiod || true
	#docker rm -f istiod || true
	docker run -it --rm  \
    -p 9090:8080 \
    -e PORT=8080 \
    -m 2G \
    --cpus=1 \
    -e K_SERVICE=dev \
    -e K_CONFIGURATION=dev \
    -e K_REVISION=dev-00001 \
    -e PROJECT=${PROJECT_ID} \
    -e CLUSTER=${CLUSTER} \
    -e ZONE=${ZONE} \
    -e REV=${REV} \
    -e LOG_ARGS=--log_output_level=info \
    -e CLOUDSDK_AUTH_CREDENTIAL_FILE_OVERRIDE=/var/run/secrets/google/google.json \
    -e GOOGLE_APPLICATION_CREDENTIALS=/var/run/secrets/google/google.json \
    -v ${GOOGLE_APPLICATION_CREDENTIALS}:/var/run/secrets/google/google.json:ro \
    gcr.io/${ISTIO_PROJECT}/pilot:${TAG}

   	#--entrypoint /bin/sh \
  #-e GODEBUG=http2client=0 \
  #  -e GOMAXPROCS=1 \

# 8 core -
# 4 core - 2 sec
# 2 core - 9 sec
# 1 core - 20 sec
#

testrun:
	kubectl run -n httpbin shell --rm -i --tty --image ubuntu -- bash

CA_ADDR ?= istiod-${PROJECT_ID}-${CLUSTER}${DOMAIN}:443
SERVICE_ACCOUNT ?= default
INTERMEDIATE_KEYSZ ?= 4096
INTERMEDIATE_ORG ?= Istio
INTERMEDIATE_CN ?= Intermediate CA

# Get a token from the project
%/token: L=$(dir $@)
%/token:
	mkdir -p ${L}
	echo '{"kind":"TokenRequest","apiVersion":"authentication.k8s.io/v1","spec":{"audiences":["istio-ca"], "expirationSeconds":2592000}}' | \
    		kubectl create --raw /api/v1/namespaces/${L})serviceaccounts/default/token -f - | \
    		jq -j '.status.token' > ${L}istio-token
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

