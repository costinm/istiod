ISTIOD?=$(shell cd ../..; pwd)
-include ${ISTIOD}/.local.mk



REPO?=gcr.io/dmeshgate

SINGLE_IMAGE?=${REPO}/proxybase:latest

FORTIO_IMAGE?=${REPO}/fortio:latest
PROJECT?=dmeshgate

#Old
#CR_SUFFIX=-yydsuf6tpq-uc.a.run.app

CR_SUFFIX=-icq63pqnqq-uc.a.run.app

SUBDOMAIN?=svc.i.webinf.info
ISTIOD_URL?=istiod.${SUBDOMAIN}:443

BASE=${ISTIOD}/samples/docker


NS=fortio

# Build the base and test docker images
docker: docker-base docker-fortio docker-iperf3

patch-fortio:
	(cd ${BASE} ; docker build . -f Dockerfile.patch \
       --build-arg BASE=fortio/fortio:latest -t ${REPO}/istio-fortio:latest)


# Base image, based on Istio proxyv2 with few additional commands
docker-base:
	docker pull gcr.io/istio-testing/proxyv2:latest
	(cd ${BASE} ; docker build . -f Dockerfile -t ${REPO}/proxybase:latest)

push-base: docker-base
	docker push ${REPO}/proxybase:latest

# k8s-run is the new 'base' - built with ko (faster)
# TODO: use distroless proxyv2 for prod
build/k8s-run: export KO_DOCKER_REPO=${REPO}/k8s-run
build/k8s-run:
	cd ${ISTIOD}/cmd/k8s-run && ko publish --bare . -t latest


docker-fortio:
	docker pull ${REPO}/k8s-run:latest
	(cd ${BASE} ; docker build . -f Dockerfile.fortio --build-arg BASE=${REPO}/k8s-run:latest -t ${FORTIO_IMAGE})

docker-iperf3:
	(cd ${BASE} ; docker build . -f Dockerfile.iperf3 --build-arg BASE=${REPO}/k8s-run:latest -t ${REPO}/iperf3:latest)

push-fortio: docker-fortio
	docker push ${FORTIO_IMAGE}

push-iperf3:
	docker push ${REPO}/iperf3:latest

build-push-base: docker-base push-base


# Common CLI args
COMMON_DOCKER=-e JWT="$(shell cat ./var/run/secrets/kubernetes.io/serviceaccount/token )" \
              -e POD_NAMESPACE=${NS} \
              -e ISTIOD=${ISTIOD_URL} \
              -e SSH_AUTH="$(shell cat ~/.ssh/id_ecdsa.pub)" \


# Deploy fortio to CloudRun
deploy: deploy-fortio


all-gw: push-base build/k8s-run deploy-gw

build-deploy-gw: build/k8s-run deploy-gw

all-fortio: push-base build/k8s-run push-fortio deploy-fortio

build-deploy-fortio: build/k8s-run push-fortio deploy-fortio

#--region asia-northeast1 \

deploy-fortio:
	gcloud alpha run deploy fortio \
		  --platform managed --project ${PROJECT} \
		  --region us-central1 \
          --sandbox=minivm \
         --allow-unauthenticated \
 		 --set-env-vars="XDS_ADDR=${ISTIOD_URL}" \
         --use-http2 \
         --port 15009 \
         --min-instances=1 \
         --image ${REPO}/fortio:latest \
 		  --set-env-vars="SSH_CA=sshca.${SUBDOMAIN}:443" \
         --set-env-vars="POD_NAMESPACE=fortio" \
 		 --set-env-vars="POD_NAME=fortio-cr" \
 		 --set-env-vars="PROJECT=${PROJECT}" \
         --set-env-vars="LOCATION=us-central1-c" \
         --set-env-vars="CLUSTER=istio" \
         --set-env-vars="LABEL_APP=fortio-cr" \
          --set-env-vars="SSH_AUTH=$(shell cat ~/.ssh/id_ecdsa.pub)" \
 		  --vpc-connector projects/wlhe-cr/locations/us-central1/connectors/serverlesscon \

# Just k8s-run and pilot-agent - used for gateway
deploy-gw:
	gcloud alpha run deploy gw \
		  --platform managed --project ${PROJECT} --region us-central1 \
          --sandbox=minivm \
         --allow-unauthenticated \
 		 --set-env-vars="XDS_ADDR=${ISTIOD_URL}" \
         --use-http2 \
         --port 15009 \
         --image ${REPO}/k8s-run:latest \
 		 --vpc-connector serverlesscon \
         --set-env-vars="SSH_CA=sshca.${SUBDOMAIN}:443" \
         --set-env-vars="POD_NAMESPACE=istio-system" \
 		 --set-env-vars="POD_NAME=istio-gateway" \
         --set-env-vars="SSH_AUTH=$(shell cat ~/.ssh/id_ecdsa.pub)" \
		 --set-env-vars="PROJECT=${PROJECT}" \
         --set-env-vars="LOCATION=us-central1-c" \
         --set-env-vars="CLUSTER=istio" \


#deploy-iperf3:
#	gcloud alpha run deploy gw ${COMMON_CR}\
#	     --image ${REPO}/iperf3:latest \
# 		 --set-env-vars="POD_NAMESPACE=fortio" \
# 		 --set-env-vars="POD_NAME=iperf3" \

# All steps up to deploy
all: docker token push deploy

# SSH to the deployed CloudRun using HBONE
ssh-fortio:
	 ssh -v  -F /dev/null -o StrictHostKeyChecking=no \
        -o "UserKnownHostsFile /dev/null" \
 	    -o ProxyCommand='hbone https://fortio-icq63pqnqq-uc.a.run.app:443/_hbone/22' \
     	root@proxybase

# Test instance in asia
ssh-fortio-asia:
	 ssh -v  -F /dev/null -o StrictHostKeyChecking=no \
        -o "UserKnownHostsFile /dev/null" \
 	    -o ProxyCommand='hbone https://fortio-icq63pqnqq-an.a.run.app:443/_hbone/22' \
     	root@proxybase

ssh-gw:
	 ssh -v  -F /dev/null -o StrictHostKeyChecking=no \
        -o "UserKnownHostsFile /dev/null" \
 	    -o ProxyCommand='hbone https://gw-icq63pqnqq-uc.a.run.app:443/_hbone/22' \
     	root@proxybase

setupcon-sharedvpc:
	gcloud services enable vpcaccess.googleapis.com

	# Requires a /28 subnet - shared vpc
	gcloud compute networks vpc-access connectors create serverlesscon \
    --region REGION \
    --subnet SUBNET \
    --subnet-project HOST_PROJECT_ID \
    --min-instances 2 \
    --max-instances 10 \

	#
	gcloud compute networks vpc-access connectors create CONNECTOR_NAME \
	--network VPC_NETWORK \
	--region REGION \
	--range IP_RANGE

sandbox:
	gcloud config set api_endpoint_overrides/run https://staging-run.sandbox.googleapis.com/

sandbox-off:
	gcloud config set api_endpoint_overrides/run https://run.googleapis.com/

istio-vm/gen:
	istioctl x workload entry configure -f /ws/istiod/samples/docker/k8s/workloadgroup-fortio.yaml -o /tmp

###################################################
# Running and testing in local docker

# Local run for the base/gateway image
# Local port 9997 forwarded to the HBONE-H2C port in the container
# (secure network mode, no mTLS)
#
# TODO: for GW, allow using 443 for mixed hbone and normal traffic
# TODO: extend GW to allow explicit HBONE config
run-base:
	docker run -it --name gw --rm \
		-p 9998:8080 -p 9997:15009\
		-e JWT="$(shell cat ./var/run/secrets/kubernetes.io/serviceaccount/token )" \
		-e POD_NAMESPACE=${NS} \
		-e ISTIOD=${ISTIOD_URL} \
		-e SSH_AUTH="$(shell cat ~/.ssh/id_ecdsa.pub)" \
		${SINGLE_IMAGE}

# Local run for the fortio image.
run-fortio:
	docker run -it --name fortio --rm \
		-p 9999:15009 \
		-e POD_NAME=localfortio \
		${COMMON_DOCKER} \
		${FORTIO_IMAGE}

run-iperf3:
	docker run -it --name iperf3 --rm \
		-p 9999:15009 \
		${COMMON_DOCKER} \
		${REPO}/iperf3:latest

# SSH to the deployed CloudRun using HBONE
local-ssh:
	 ssh -F /dev/null -o "UserKnownHostsFile /dev/null" \
 		-o StrictHostKeyChecking=no \
 		-o ProxyCommand='hbone http://127.0.0.1:9997/_hbone/22'  \
     		root@proxybase

local-ssh-fortio:
	 ssh -F /dev/null -o "UserKnownHostsFile /dev/null" \
 		-o StrictHostKeyChecking=no \
 		-o ProxyCommand='hbone http://127.0.0.1:9999/_hbone/22'  \
     		root@proxybase


# Get a JWT token - will be used to authenticate (until OIDC is fixed)
token: SECRET_NAME=$(shell kubectl get sa default -n ${NS} -o jsonpath='{.secrets[].name}')
token: TOKEN=$(shell kubectl get secret ${SECRET_NAME} -n ${NS} -o jsonpath="{.data['token']}")
token:
	mkdir -p ./var/run/secrets/kubernetes.io/serviceaccount/
	echo ${TOKEN} |  base64 --decode > ./var/run/secrets/kubernetes.io/serviceaccount/token
