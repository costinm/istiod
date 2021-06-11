#REPO ?= costinm
REPO?=gcr.io/dmeshgate
SINGLE_IMAGE?=${REPO}/proxybase:latest
FORTIO_IMAGE?=${REPO}/fortio:latest
PROJECT?=dmeshgate
SVC?=proxybase-yydsuf6tpq-uc.a.run.app

ISTIOD?=$(shell cd ../..; pwd)
BASE=${ISTIOD}/samples/knative

NS=fortio

# Build the base and test docker images
single/docker:
	#(cd ${BASE} ; docker build . -f Dockerfile -t ${SINGLE_IMAGE})
	(cd ${BASE} ; docker build . -f Dockerfile.fortio --build-arg BASE=${SINGLE_IMAGE} -t ${FORTIO_IMAGE})

single/docker-base:
	(cd ${BASE} ; docker build . -f Dockerfile -t ${SINGLE_IMAGE})

single/docker-fortio:
	(cd ${BASE} ; docker build . -f Dockerfile.fortio --build-arg BASE=${SINGLE_IMAGE} -t ${FORTIO_IMAGE})

# Push the images
single/push:
	#docker push ${SINGLE_IMAGE}
	docker push ${FORTIO_IMAGE}

# Get a JWT token - will be used to authenticate (until OIDC is fixed)
single/token: SECRET_NAME=$(shell kubectl get sa default -n ${NS} -o jsonpath='{.secrets[].name}')
single/token: TOKEN=$(shell kubectl get secret ${SECRET_NAME} -n ${NS} -o jsonpath="{.data['token']}")
single/token:
	mkdir -p ./var/run/secrets/kubernetes.io/serviceaccount/
	echo ${TOKEN} |  base64 --decode > ./var/run/secrets/kubernetes.io/serviceaccount/token

# Local run for the docker image
single/run-base:
	docker run -it --name fortio --rm \
		-p 9998:8080 -p 9997:15002\
		-e JWT="$(shell cat ./var/run/secrets/kubernetes.io/serviceaccount/token )" \
		-e POD_NAMESPACE=${NS} \
		-e ISTIOD=k.webinf.info:443 \
		-e SSH_AUTH="$(shell cat ~/.ssh/id_ecdsa.pub)" \
		${SINGLE_IMAGE}

single/run-fortio:
	docker run -it --name fortio --rm \
		-p 9999:15002 \
    	-e JWT="$(shell cat ./var/run/secrets/kubernetes.io/serviceaccount/token )" \
		-e POD_NAMESPACE=${NS} \
		-e ISTIOD=k.webinf.info:443 \
		-e SSH_AUTH="$(shell cat ~/.ssh/id_ecdsa.pub)" \
		${FORTIO_IMAGE}

# Deploy fortio to CloudRun
single/deploy:
	#--command /usr/local/bin/run.sh \
	gcloud alpha run deploy fortio  --platform managed --project ${PROJECT} \
		 --sandbox=minivm \
		 --port 15002 \
 		--region us-central1 --image ${FORTIO_IMAGE} \
 		--command "/usr/local/bin/run.sh" --args "/usr/bin/fortio,server,-http-port=8082" \
 		 --allow-unauthenticated \
 		 --use-http2 \
 		 --set-env-vars="POD_NAMESPACE=fortio" \
 		 --set-env-vars="POD_NAME=fortiod" \
 		 --set-env-vars="ISTIOD=k.webinf.info:443" \
         --set-env-vars="SSH_AUTH=$(shell cat ~/.ssh/id_ecdsa.pub)" \
		 --set-env-vars="JWT=$(shell cat ./var/run/secrets/kubernetes.io/serviceaccount/token)"

# All steps up to deploy
single/all: single/docker single/token single/push single/deploy

# SSH to the deployed CloudRun using HBONE
single/ssh:
	 ssh -v  -F /dev/null -o StrictHostKeyChecking=no \
       -o ProxyCommand='hbone https://${SVC}:443/dm/127.0.0.1:22'  \
     		root@proxybase

# SSH to the deployed CloudRun using HBONE
single/local-ssh:
	 ssh -F /dev/null -o StrictHostKeyChecking=no \
       -o ProxyCommand='hbone https://127.0.0.1:9997/hbone/22'  \
     		root@proxybase
