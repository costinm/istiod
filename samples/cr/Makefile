# Run
# export HUB=gcr.io/wlhe-cr
# export TAG=asm-cr
#  make push.docker.cloudrun push.docker.proxyv2

ISTIO_REGION ?= us-central1
ISTIO_PROJECT ?= wlhe-cr
ISTIOD_SA ?= 601426346923-compute@developer.gserviceaccount.com
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

# asm-cr should only be built out of git-on-borg !!!
TAG ?= asm-canary

HUB ?= gcr.io/${ISTIO_PROJECT}
# TODO: export the image on wlhe
PROXY_HUB ?= ${HUB}
export TAG
export HUB
export REV
export VERSION=1.8-$(shell date +%y%m%d)
# Suffix for the domain URLs -
ISTIO_SUFFIX?=-icq63pqnqq-uc.a.run.app
#SISTIO_SUFFIX=-t2mbbtjbiq-uc.a.cloud-run-qual.sandbox.google.com
URL=istiod-${PROJECT_ID}-${CLUSTER}-${REV}${ISTIO_SUFFIX}

# Default to the docker env dir
ISTIO_GO ?= /work
export TARGET ?= -asm
ISTIO_OUT=${ISTIO_GO}/out/
DOCKER_BUILD_TOP:=${ISTIO_OUT}/docker_build

include test.mk
include clusters.mk

# Main testing target for ASM
all: knative/cloudrun knative/canary knative/prober

build: knative/cloudrun knative/proxy

gob:
	TAG=asm-gob REV=asm-gob ENVEXTRA="ASM=1," $(MAKE) _run
	TAG=asm-gob REV=asm-gob $(MAKE) fortio

# Common target for running
_run: PROJECT_NUMBER=$(shell gcloud projects describe ${PROJECT_ID} --format="value(projectNumber)")
_run:
	gcloud alpha run deploy istiod-${PROJECT_ID}-${CLUSTER}-${REV} --allow-unauthenticated \
	 --project ${ISTIO_PROJECT} --region ${ISTIO_REGION} --platform managed \
     --image gcr.io/${ISTIO_PROJECT}/cloudrun:${TAG} \
     --set-env-vars=HUB=${HUB},${ENVEXTRA}REV=${REV},PROJECT_NUMBER=${PROJECT_NUMBER},TAG=${TAG},CLUSTER=${CLUSTER},ZONE=${ZONE},PROJECT=${PROJECT_ID},ISTIOD_DOMAIN=${ISTIO_SUFFIX} \
	 --port 8080 --concurrency 1000 --timeout 900 --cpu ${CPU} --memory ${MEM}

run3: PROJECT_NUMBER=$(shell gcloud projects describe ${PROJECT_ID} --format="value(projectNumber)")
run3:
	gcloud alpha run deploy istiod-${PROJECT_ID}-${CLUSTER}-asm --allow-unauthenticated \
	 --project ${ISTIO_PROJECT} --region ${ISTIO_REGION} --platform managed \
     --image gcr.io/${ISTIO_PROJECT}/cloudrun:${TAG} \
     --set-env-vars=${ENVEXTRA}REV=asm,ASM=1,PROJECT_NUMBER=${PROJECT_NUMBER},TAG=${TAG},CLUSTER=${CLUSTER},ZONE=${ZONE},PROJECT=${PROJECT_ID},ISTIOD_DOMAIN=${ISTIO_SUFFIX} \
	 --port 8080 --concurrency 1000 --timeout 900 --cpu ${CPU} --memory ${MEM} --min-instances=1
	gcloud alpha run deploy istiod-${PROJECT_ID}-${CLUSTER}-ossmanaged --allow-unauthenticated \
	 --project ${ISTIO_PROJECT} --region ${ISTIO_REGION} --platform managed \
     --image gcr.io/${ISTIO_PROJECT}/cloudrun:${TAG} \
     --set-env-vars=${ENVEXTRA}REV=ossmanaged,PROJECT_NUMBER=${PROJECT_NUMBER},TAG=${TAG},CLUSTER=${CLUSTER},ZONE=${ZONE},PROJECT=${PROJECT_ID},ISTIOD_DOMAIN=${ISTIO_SUFFIX} \
	 --port 8080 --concurrency 1000 --timeout 900 --cpu ${CPU} --memory ${MEM}
	gcloud alpha run deploy istiod-${PROJECT_ID}-${CLUSTER}-asmca --allow-unauthenticated \
	 --project ${ISTIO_PROJECT} --region ${ISTIO_REGION} --platform managed \
     --image gcr.io/${ISTIO_PROJECT}/cloudrun:${TAG} \
     --set-env-vars=${ENVEXTRA}REV=asmca,ASM=1,CA=1,PROJECT_NUMBER=${PROJECT_NUMBER},TAG=${TAG},CLUSTER=${CLUSTER},ZONE=${ZONE},PROJECT=${PROJECT_ID},ISTIOD_DOMAIN=${ISTIO_SUFFIX} \
	 --port 8080 --concurrency 1000 --timeout 900 --cpu ${CPU} --memory ${MEM}


knative/cloudrun:
	(cd ${ISTIO_GO} ; TAG=${TAG} HUB=${HUB} make push.docker.cloudrun)

# Build the proxy in a separate hub (until wlhe repo is public)
knative/proxy:
	cd ${ISTIO_GO} && HUB=${PROXY_HUB} TAG=${TAG} make push.docker.proxyv2

# Run just the ASM version.
knative/canary:
	REV=asm ENVEXTRA="ASM=1," $(MAKE) _run
	REV=oss $(MAKE) _run

knative/prober:
	REV=asm $(MAKE) fortio
	REV=oss $(MAKE) fortio


# Restart the fortio images in PROJECT_ID / CLUSTER / REV.
knative/test/proxy:
	#kubectl --context gke_wlhe-cr_us-central1-c_istio -n fortio-${REV} delete po --all --force --grace-period=0
	#kubectl --context gke_costin-asm1_us-central1-c_cloudrun -n fortio-${REV} delete po --all --force --grace-period=0
	kubectl --context gke_costin-asm1_us-central1-c_big1 -n fortio-${REV} delete po --all --force --grace-period=0

# Steps:
# - create cluster
# - grant IAM to cloud run to cluster
# - base install
   	#--entrypoint /bin/sh \
  #-e GODEBUG=http2client=0 \
  #  -e GOMAXPROCS=1 \

GOOGLE_APPLICATION_CREDENTIALS=${HOME}/.config/gcloud/legacy_credentials/costin@google.com/adc.json
local/docker:
	#docker stop istiod || true
	#docker rm -f istiod || true
	docker run -it --rm  \
    -p 9091:8080 \
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
    -e ASM=1 \
    -e XDS_ADDR=localhost:9091 \
    -e LOG_ARGS=--log_output_level=info \
    -e CLOUDSDK_AUTH_CREDENTIAL_FILE_OVERRIDE=/var/run/secrets/google/google.json \
    -e GOOGLE_APPLICATION_CREDENTIALS=/var/run/secrets/google/google.json \
    -v ${GOOGLE_APPLICATION_CREDENTIALS}:/var/run/secrets/google/google.json:ro \
    gcr.io/${ISTIO_PROJECT}/cloudrun:${TAG}

   	#--entrypoint /bin/sh \
  #-e GODEBUG=http2client=0 \
  #  -e GOMAXPROCS=1 \

# Temp: Init the current project for staging
gcp/init:
	gcloud services enable staging-run.sandbox.googleapis.com
	gcloud config set run/region us-central1
	gcloud config set run/platform managed
	gcloud config set project costin-asm1

	 gcloud container clusters update big1 --zone us-central1-c --enable-stackdriver-kubernetes
	#gcloud config unset api_endpoint_overrides/run
	#https://gce-staging.pantheon-ci.corp.google.com/run?project=costin-asm1
	#gcloud config set api_endpoint_overrides/run https://staging-run.sandbox.googleapis.com/

cr/status:
	gcloud run services --platform managed list

# --concurrency= = 0 unspec, set a max per instance
# --service-account - create one per tenant
# --timeout - 10m5s

TOP=$(shell pwd)

build/telemetry-gen:
	(cd ${ISTIO_GO} && helm3 template -n istio-system istio-default \
		manifests/charts/istio-control/istio-discovery \
		--set telemetry.v2.stackdriver.enabled=true --set telemetry.v2.prometheus.enabled=true \
        --set telemetry.v2.prometheus.wasmEnabled=true --set telemetry.v2.accessLogPolicy.enabled=true \
        --set revision=XREVISION --set telemetry.v2.metadataExchange.wasmEnabled=true  \
        --set global.multiCluster.clusterName=XCLUSTER ) | sed s/XREVISION/\$${REVISION}/ > istio-all.yaml

#gcloud config unset api_endpoint_overrides/run
#export DOCKER_HOST=tcp://localhost:2375

# Run 2 revisions - one ASM one OSS
# The OSS revision can be used in non-GKE clusters
# The ASM will use ManagedCA and Stackdriver, on GKE
run:
	REV=asm ENVEXTRA="ASM=1," CPU=4 MEM=4G $(MAKE) _run

#LOG_ARGS=--log_output_level=debug,
#gcloud config unset api_endpoint_overrides/run
#--max-instances=1
#--cpu 4 --memory 4G --min-instances=1

# Max set to 1 to test single-instance perf
#gcloud config set api_endpoint_overrides/run https://staging-run.sandbox.googleapis.com/
#gcloud config unset api_endpoint_overrides/run

staging: build/staging run_staging

push: build/remote run run_staging

logs:
	 gcloud logging read 'resource.type = "project" OR resource.type = "cloud_run_revision"'


xds:
	echo '{"node": {"id": "sidecar~1.1.1.1~debug~cluster.local"}, "typeUrl":"type.googleapis.com/envoy.config.cluster.v3.Cluster"}' | \
		grpcurl -d @ istiod-costin-asm1-big1-asm-canary-icq63pqnqq-uc.a.run.app:443  envoy.service.discovery.v3.AggregatedDiscoveryService/StreamAggregatedResources

cert:
	echo {} | grpcurl  -import-path /work/security/proto/ -proto /work/security/proto/istioca.proto -use-reflection=false -v -d @ -H "Authorization:  Bearer $(gcloud auth print-identity-token)" -rpc-header "authorization: Bearer $(gcloud auth print-identity-token)"  istiodxds-b3lo6zbfmq-uc.a.run.app:443 istio.v1.auth.IstioCertificateService/CreateCertificate


# Verify injection is active
test/inject:
	time curl  https://istiod-${PROJECT_ID}-${CLUSTER}-${REV}${ISTIO_SUFFIX}/inject  -d @test/mutating_req.json -HContent-Type:application/json | jq .response.patch | sed s/\"// | base64 -d | jq .


test/dump:
	curl -H "Authorization: Bearer $(gcloud auth print-identity-token)" https://${URL}:443/debug/configz

test/ep:
	curl -H "Authorization: Bearer $(gcloud auth print-identity-token)" https://${URL}:443/debug/connectionz


test/grpc:
	echo '{"node": {"id": "sidecar~1.1.1.1~debug~cluster.local"},"typeUrl": "istio.io/connections"}' \
      | grpcurl -v -d @ ${URL}:443 envoy.service.discovery.v3.AggregatedDiscoveryService/StreamAggregatedResources


# Inject on a local workload
local/inject:
	curl  http://localhost:8080/inject/  -d @mutating_req.json -HContent-Type:application/json  | jq .response.patch | sed s/\"// | base64 -d | jq .

test/ps: INJECT_GRPC=$(shell kubectl get mutatingwebhookconfiguration istiod-asm-managed -o json | jq -r .webhooks[0].clientConfig.url | sed s,/inject.*,, | sed s,https://,, )
test/ps:
	${ISTIO_GO}/out/linux_amd64/istioctl x ps --xds-address ${INJECT_GRPC}


# 8 core -
# 4 core - 2 sec
# 2 core - 9 sec
# 1 core - 20 sec
#

testrun:
	kubectl run -n httpbin shell --rm -i --tty --image ubuntu -- bash

knative/project_setup:
	gcloud projects add-iam-policy-binding \
            ${PROJECT_ID} \
            --member="serviceAccount:${ISTIOD_SA}" \
            --role="roles/container.admin"
	gcloud --project ${PROJECT_ID} services enable \
        container.googleapis.com \
        compute.googleapis.com \
        monitoring.googleapis.com \
        logging.googleapis.com \
        cloudtrace.googleapis.com \
        meshca.googleapis.com \
        meshtelemetry.googleapis.com \
        meshconfig.googleapis.com \
        iamcredentials.googleapis.com \
        anthos.googleapis.com \
        gkeconnect.googleapis.com \
        gkehub.googleapis.com \
        cloudresourcemanager.googleapis.com

#  gcloud projects add-iam-policy-binding wlhe-cr --member="serviceAccount:601426346923-compute@developer.gserviceaccount.com" --role="roles/serviceusage.serviceUsageConsumer"

knative/asm_init:
	curl --request POST \
      --header "Authorization: Bearer $(shell gcloud --project ${PROJECT_ID} auth print-access-token)" \
      --data '' \
    "https://meshconfig.googleapis.com/v1alpha1/projects/${PROJECT_ID}:initialize"

managed/gateway:
	helm3  -n istio-system template  manifests/charts/gateways/istio-ingress \
        -f manifests/charts/global.yaml \
		--set global.pilotCertProvider=system \
		--set-string meshConfig.defaultConfig.proxyMetadata.PROXY_CONFIG='\{"discoveryAddress":"foo.bar"\}'

