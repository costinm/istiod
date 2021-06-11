#REPO ?= costinm
REPO?=gcr.io/dmeshgate
SINGLE_IMAGE?=${REPO}/proxybase:latest
FORTIO_IMAGE?=${REPO}/fortio:latest
PROJECT?=dmeshgate
CR_SUFFIX=-yydsuf6tpq-uc.a.run.app

ISTIOD?=$(shell cd ../..; pwd)
BASE=${ISTIOD}/samples/knative

NS=fortio

# Build the base and test docker images
single/docker: single/docker-base single/docker-fortio single/docker-iperf3

single/docker-base:
	(cd ${BASE} ; docker build . -f Dockerfile -t ${SINGLE_IMAGE})

single/docker-fortio:
	(cd ${BASE} ; docker build . -f Dockerfile.fortio --build-arg BASE=${SINGLE_IMAGE} -t ${FORTIO_IMAGE})

single/docker-iperf3:
	(cd ${BASE} ; docker build . -f Dockerfile.iperf3 --build-arg BASE=${SINGLE_IMAGE} -t ${REPO}/iperf3:latest)

# Push the images
single/push: single/push-base single/push-fortio single/push-iperf3

single/push-fortio:
	docker push ${FORTIO_IMAGE}

single/push-iperf3:
	docker push ${REPO}/iperf3:latest

single/push-base:
	docker push ${SINGLE_IMAGE}

# Get a JWT token - will be used to authenticate (until OIDC is fixed)
single/token: SECRET_NAME=$(shell kubectl get sa default -n ${NS} -o jsonpath='{.secrets[].name}')
single/token: TOKEN=$(shell kubectl get secret ${SECRET_NAME} -n ${NS} -o jsonpath="{.data['token']}")
single/token:
	mkdir -p ./var/run/secrets/kubernetes.io/serviceaccount/
	echo ${TOKEN} |  base64 --decode > ./var/run/secrets/kubernetes.io/serviceaccount/token

# Common CLI args
COMMON_DOCKER=-e JWT="$(shell cat ./var/run/secrets/kubernetes.io/serviceaccount/token )" \
              -e POD_NAMESPACE=${NS} \
              -e ISTIOD=k.webinf.info:443 \
              -e SSH_AUTH="$(shell cat ~/.ssh/id_ecdsa.pub)" \


# Local run for the base/gateway image
# Local port 9997 forwarded to the HBONE-H2C port in the container
# (secure network mode, no mTLS)
#
# TODO: for GW, allow using 443 for mixed hbone and normal traffic
# TODO: extend GW to allow explicit HBONE config
single/run-base:
	docker run -it --name gw --rm \
		-p 9998:8080 -p 9997:15009\
		-e JWT="$(shell cat ./var/run/secrets/kubernetes.io/serviceaccount/token )" \
		-e POD_NAMESPACE=${NS} \
		-e ISTIOD=k.webinf.info:443 \
		-e SSH_AUTH="$(shell cat ~/.ssh/id_ecdsa.pub)" \
		${SINGLE_IMAGE}

# Local run for the fortio image.
single/run-fortio:
	docker run -it --name fortio --rm \
		-p 9999:15009 \
		-e POD_NAME=localfortio \
		${COMMON_DOCKER} \
		${FORTIO_IMAGE}

single/run-iperf3:
	docker run -it --name iperf3 --rm \
		-p 9999:15009 \
		${COMMON_DOCKER} \
		${REPO}/iperf3:latest

# Deploy fortio to CloudRun
single/deploy: single/deploy-fortio

COMMON_CR=--platform managed --project ${PROJECT} \
          		 --sandbox=minivm \
         --allow-unauthenticated \
 		 --set-env-vars="ISTIOD=k.webinf.info:443" \
          		 --use-http2 \
          		  --port 15009 \
         --set-env-vars="SSH_AUTH=$(shell cat ~/.ssh/id_ecdsa.pub)" \
		 --set-env-vars="JWT=$(shell cat ./var/run/secrets/kubernetes.io/serviceaccount/token)" \
           		--region us-central1

single/deploy-fortio:
	#--command "/usr/local/bin/run.sh" --args "/usr/bin/fortio,server,-http-port=8082" \
    #--command /usr/local/bin/run.sh
	gcloud alpha run deploy fortio  --platform managed --project ${PROJECT} \
		 --sandbox=minivm \
		 --port 15009 \
    	 --region us-central1 --image ${FORTIO_IMAGE} \
 		 --allow-unauthenticated \
 		 --use-http2 \
 		 --set-env-vars="POD_NAMESPACE=fortio" \
 		 --set-env-vars="POD_NAME=fortiod" \
 		 --set-env-vars="ISTIOD=k.webinf.info:443" \
         --set-env-vars="SSH_AUTH=$(shell cat ~/.ssh/id_ecdsa.pub)" \
		 --set-env-vars="JWT=$(shell cat ./var/run/secrets/kubernetes.io/serviceaccount/token)"

single/deploy-gw:
	gcloud alpha run deploy gw ${COMMON_CR}\
	     --image ${SINGLE_IMAGE} \
 		 --set-env-vars="POD_NAMESPACE=fortio" \
 		 --set-env-vars="POD_NAME=istio-gateway" \

single/deploy-iperf3:
	gcloud alpha run deploy gw ${COMMON_CR}\
	     --image ${REPO}/iperf3:latest \
 		 --set-env-vars="POD_NAMESPACE=fortio" \
 		 --set-env-vars="POD_NAME=iperf3" \

# All steps up to deploy
single/all: single/docker single/token single/push single/deploy

# SSH to the deployed CloudRun using HBONE
single/ssh-fortio:
	 ssh -v  -F /dev/null -o StrictHostKeyChecking=no \
        -o "UserKnownHostsFile /dev/null" \
 	    -o ProxyCommand='hbone https://fortio${CR_SUFFIX}:443/_hbone/22' \
     	root@proxybase

# SSH to the deployed CloudRun using HBONE
single/local-ssh:
	 ssh -F /dev/null -o "UserKnownHostsFile /dev/null" \
 		-o StrictHostKeyChecking=no \
 		-o ProxyCommand='hbone http://127.0.0.1:9997/_hbone/22'  \
     		root@proxybase

single/local-ssh-fortio:
	 ssh -F /dev/null -o "UserKnownHostsFile /dev/null" \
 		-o StrictHostKeyChecking=no \
 		-o ProxyCommand='hbone http://127.0.0.1:9999/_hbone/22'  \
     		root@proxybase
