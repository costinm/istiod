ISTIO_REGION ?= us-central1
ISTIO_PROJECT ?= wlhe-cr

PROJECT ?= costin-asm1
ZONE ?= us-central1-c
CLUSTER ?= big1
TAG ?= cr
HUB ?= gcr.io/${ISTIO_PROJECT}
export TAG
export HUB

# Suffix for the domain URLs -
DOMAIN=-icq63pqnqq-uc.a.run.app
#SDOMAIN=-t2mbbtjbiq-uc.a.cloud-run-qual.sandbox.google.com

init:
	gcloud services enable staging-run.sandbox.googleapis.com

	gcloud config set run/region us-central1
	gcloud config set run/platform managed
	gcloud config set project costin-asm1

	gcloud config set api_endpoint_overrides/run https://staging-run.sandbox.googleapis.com/

runstatus:
	gcloud run services --platform managed list

deploy:
	docker pull costinm/pilot:latest
	docker tag costinm/pilot:latest gcr.io/costin-asm1/pilot:latest
	docker push gcr.io/costin-asm1/pilot:latest



run_old:
	#gcloud run deploy istiod --allow-unauthenticated --image gcr.io/costin-asm1/pilot:latest --args=discovery,--log_output_level=debug

	gcloud run deploy istiod --allow-unauthenticated --image gcr.io/costin-asm1/pilot:latest \
     --args=discovery,--log_output_level=debug,--registries=Mock,--configDir=/var/lib/istio \
      curl -v  https://istiod-xds-icq63pqnqq-uc.a.run.app/debug/configz \
	--port 15010
# Command - to override the default command ( pilot )

# --concurrency= = 0 unspec, set a max per instance
# --cpu
# --memory
# --max-instances
# --port=8080 - also sets PORT
# --service-account - create one per tenant
# --timeout - 10m5s
#https://gce-staging.pantheon-ci.corp.google.com/run?project=costin-asm1

dump:
	curl -H "Authorization: Bearer $(gcloud auth print-identity-token)" https://istiod-b3lo6zbfmq-uk.a.run.app/debug/configz

test:
	 grpcurl -v -d @ istiod-xds-icq63pqnqq-uc.a.run.app:443 envoy.service.discovery.v3.AggregatedDiscoveryService/StreamAggregatedResources

#{"node": {"id": "sidecar~1.1.1.1~debug~cluster.local", "metadata": {"GENERATOR": "event"}},"typeUrl": "istio.io/connections"}

TOP=$(shell pwd)

build/istiod:
	(cd /work; go build -o ${TOP}/istiod ./pilot/cmd/pilot-discovery)


#export DOCKER_HOST=tcp://localhost:2375

build/remote: build/istiod
	docker build -t ${HUB}/pilot:${TAG} .
	docker push  ${HUB}/pilot:${TAG}

push/proxy:
	cd /work && HUB=${HUB} TAG=${TAG} make push.docker.proxyv2
	docker tag ${HUB}/proxyv2:${TAG} costinm/proxyv2:${TAG}
	docker push costinm/proxyv2:${TAG}

run:
	#gcloud config unset api_endpoint_overrides/run
	make _run

_run:
	#gcloud config unset api_endpoint_overrides/run
	# Slow start - but pinned
	# Alternative: --cpu 4 --memory 4G
	gcloud alpha run deploy istiod${REV}-${PROJECT}-${CLUSTER} --allow-unauthenticated \
	 --project ${ISTIO_PROJECT} --region ${ISTIO_REGION} --platform managed \
     --image gcr.io/${ISTIO_PROJECT}/pilot:${TAG} \
     --set-env-vars=CLUSTER=${CLUSTER},ZONE=${ZONE},PROJECT=${PROJECT},ISTIOD_DOMAIN=${DOMAIN},EXTRA_ARGS=--grpcAddr=OFF \
	 --port 8080 --cpu 4 --memory 4G --min-instances=1

	#gcloud config set api_endpoint_overrides/run https://staging-run.sandbox.googleapis.com/
#	gcloud alpha run deploy istiod${REV}-${PROJECT}-${CLUSTER}-xds --allow-unauthenticated \
#	 --project ${ISTIO_PROJECT} --region ${ISTIO_REGION} --platform managed \
#        --image gcr.io/${ISTIO_PROJECT}/pilot:${TAG} \
#       --set-env-vars=CLUSTER=${CLUSTER},ZONE=${ZONE},PROJECT=${PROJECT},ISTIOD_DOMAIN=${DOMAIN},INJECT_ENABLED=false,VALIDATION_ENABLED=false \
#        --port 15010 --cpu 4 --memory 4G --min-instances=1
	#gcloud config unset api_endpoint_overrides/run

build/staging:
	PROJECT=costin-asm1 make build/remote


staging: build/staging run_staging

push: build/remote run run_staging

logs:
	 gcloud logging read 'resource.type = "project" OR resource.type = "cloud_run_revision"'

cert:
	echo {} | grpcurl  -import-path /work/security/proto/ -proto /work/security/proto/istioca.proto -use-reflection=false -v -d @ -H "Authorization:  Bearer $(gcloud auth print-identity-token)" -rpc-header "authorization: Bearer $(gcloud auth print-identity-token)"  istiodxds-b3lo6zbfmq-uc.a.run.app:443 istio.v1.auth.IstioCertificateService/CreateCertificate


# Verify injection is active
test/inject:
	time curl -v https://istiod${REV}-${PROJECT}-${CLUSTER}${DOMAIN}/inject  -d @mutating_req.json -HContent-Type:application/json | jq .response.patch | sed s/\"// | base64 -d | jq .



# Inject on a local workload
local/inject:
	curl  http://localhost:8080/inject/  -d @mutating_req.json -HContent-Type:application/json  | jq .response.patch | sed s/\"// | base64 -d | jq .

local/grpc:


# Steps:
# - create cluster
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
    -e PROJECT=${PROJECT} \
    -e CLUSTER=${CLUSTER} \
    -e ZONE=${ZONE} \
    -e LOG_ARGS=--log_output_level=info \
    -e CLOUDSDK_AUTH_CREDENTIAL_FILE_OVERRIDE=/var/run/secrets/google/google.json \
    -e GOOGLE_APPLICATION_CREDENTIALS=/var/run/secrets/google/google.json \
    -v ${GOOGLE_APPLICATION_CREDENTIALS}:/var/run/secrets/google/google.json:ro \
    gcr.io/${ISTIO_PROJECT}/pilot:gcp

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
