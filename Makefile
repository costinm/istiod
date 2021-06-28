# Makefile has additional test and build targets. The primary mechansims to build are:
# - 'docker build . -t costinm/istiod:latest'
# - 'go get github.com/costinm/istiod/cmd/istiod'
# - from IDE using normal run/debug command

ISTIOD=$(shell cd .; pwd)
ISTIO_SRC=$(shell cd ../istio; pwd)
TOP=$(shell cd ${BASE}/../../..; pwd)

BASE=$(shell cd .; pwd)
GOPATH=${HOME}/go

-include .local.mk

CONF ?= ${BASE}/conf
HUB ?= gcr.io/dmeshgate
#HUB ?= costinm
#HUB ?= localhost:5000
TAG ?= latest

IMAGE ?= ${HUB}/istiod

CACHEDIR ?= ${TOP}/out/cache

LOG_DIR ?= /tmp

OUT ?= ${ISTIO_SRC}/out

# Namespace to use for the test app
NS=fortio

IP ?= $(shell hostname --ip-address)

# Namespace used for testing
NAMESPACE ?= ugate

# Set to "-it --rm" to run the docker images in foreground, for testing.
# Default is to set the images as daemon.
# Set to "run -it --rm " for debugging
DOCKER_START ?= run -d

BINDIR=${OUT}/linux_amd64

build/k8s-run: export KO_DOCKER_REPO=${HUB}/k8s-run
build/k8s-run:
	cd cmd/k8s-run && ko publish --bare . -t latest

build/istiod:
	cd ${ISTIO_SRC} && CGO_ENABLED=0 \
	go build -a -ldflags '-extldflags "-static"' -o ${BINDIR}/pilot-discovery ./pilot/cmd/pilot-discovery

build/istio-agent:
	cd ${ISTIO_SRC} && \
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ${BINDIR}/pilot-agent ./pilot/cmd/pilot-agent

# Example using external OIDC
#TOKEN_ISSUER=https://container.googleapis.com/v1/projects/costin-istio/locations/us-west1-c/clusters/istio-test

run/istiod:
	cd ${ISTIO_SRC} && ${BINDIR}/pilot-discovery discovery

# Fetch bootstrap token and root cert
# John's script:
bootstrap/short:
	mkdir -p ${ISTIO_SRC}/var/run/secrets/tokens ${ISTIO_SRC}/var/run/secrets/istio
	echo '{"kind":"TokenRequest","apiVersion":"authentication.k8s.io/v1","spec":{"audiences":["istio-ca"], "expirationSeconds":2592000}}' | \
    		kubectl create --raw /api/v1/namespaces/${NAMESPACE}/serviceaccounts/default/token -f - | \
    		jq -j '.status.token' > ${ISTIO_SRC}/var/run/secrets/tokens/istio-token
	kubectl -n istio-system get secret istio-ca-secret -ojsonpath='{.data.ca-cert\.pem}' | \
      	base64 -d > ${ISTIO_SRC}/var/run/secrets/istio/root-cert.pem


PROXY_CONFIG = {"binaryPath": "${BINDIR}/release/envoy", "configPath": "${BINDIR}", "proxyBootstrapTemplatePath": \
	"${ISTIO_SRC}/tools/packaging/common/envoy_bootstrap.json", "discoveryAddress": "localhost:15012", \
    "terminationDrainDuration": "0s"}

export PROXY_CONFIG

run/sidecar:
	echo PROXY_CONFIG=${PROXY_CONFIG}
	cd ${ISTIO_SRC} && XDS_LOCAL=127.0.0.1:15002 \
		${BINDIR}/pilot-agent proxy sidecar

run/gateway:
	cd ${ISTIO_SRC} && ISTIO_META_CLUSTER_ID=Kubernetes \
	ISTIO_METAJSON_LABELS="{\"istio\": \"ingressgateway\", \"app\": \"istio-ingressgateway\"}" \
		${BINDIR}/pilot-agent proxy router


# Doesn't work with alpine
build-local-docker: istiod
	docker build ${TOP}/istiod -f tools/local_docker/Dockerfile -t ${IMAGE}:${TAG}

build-docker:
	DOCKER_BUILDKIT=1 docker build . -t ${IMAGE}:latest
	#DOCKER_BUILDKIT=1 docker build . --target distroless -t ${IMAGE}-distroless:${TAG}

push-docker:
	docker push ${IMAGE}:latest
	#docker push ${IMAGE}-distroless:latest

# Example of starting pilot standalone - replaced by istiod-vm, using galley file source
#
## Start pilot in a docker container, using a local set of files, no k8s used.
## If configDir is specified, it will be used as a direct source of config, instead of CRDs. Will skip creation
## of the kubeClient as long as 'registries' doesn't include k8s.
##
## "--registries" still needed, to disable k8s registry - even if MCP is not configured in mesh.yaml
## ServiceEntries will be loaded from the config dir.
##
## Plugins: authz, authn, mixer, health
#pilot:
#	docker rm -f pilot || true
#	docker ${DOCKER_START} --name=pilot  \
#		-p 127.0.0.1:15080:8080 \
#		-p 0.0.0.0:15010:15010 \
#		-p 127.0.0.1:15014:15014 \
#		-p 127.0.0.1:15876:9876 \
#        -v ${PWD}/conf/pilot:/var/lib/istio/pilot \
#		-v ${PWD}/conf/istio:/var/lib/istio/istio \
#		-e PILOT_ENABLE_PROTOCOL_SNIFFING=true \
#	 ${HUB}/pilot:${TAG} \
#    	 discovery --meshConfig /var/lib/istio/pilot/mesh.yaml \
#    	--secureGrpcAddr="" \
#    	--plugins="authz" \
#        --configDir /var/lib/istio/istio --registries=MCP \
#        --networksConfig /var/lib/istio/pilot/meshNetworks.yaml

# Start a local gateway, running in docker. Uses upstream envoy
#
# We need to pass the pilot address - either as an /etc/host set to host IP (or another addr),
# or as 127.0.0.1 and use the network of the pilot container.
gateway:
	docker stop gateway || true
	docker ${DOCKER_START} --name=gateway  \
		--add-host istio-pilot.istio-system:${IP} \
        -p 127.0.0.1:16080:80 \
		-p 127.0.0.1:16443:443 \
		-p 127.0.0.1:16000:15000 \
        -v ${PWD}/conf/gateway:/var/lib/istio/envoy \
	 envoyproxy/envoy-alpine-debug-dev \
	         -c /var/lib/istio/envoy/bootstrap.yaml \
             --base-id 4 \
             --log-level debug \
             --restart-epoch 0 \
             --drain-time-s 45 \
             --parent-shutdown-time-s 60 \
             --service-cluster istio-ingressgateway \
             --service-node router~10.244.1.82~istio-ingressgateway-7d467cd559-qrsbv.istio-micro~istio-micro.svc.cluster.local \
             --max-obj-name-len 189 \
             --local-address-ip-version v4 \
             --allow-unknown-fields

# Start fortio application on the local machine, with sidecar and iptables initialization
start-fortio:
	$(MAKE) fortio-sidecar
	$(MAKE) fortio-iptables
	$(MAKE) fortio-app

stop-fortio:
	docker stop fortio-sidecar || true
	docker stop fortio-app || true
	docker rm fortio-sidecar || true
	docker rm fortio-app || true

fortio-sidecar:
	docker stop fortio-sidecar || true
	docker rm fortio-sidecar || true
	docker ${DOCKER_START} --name=fortio-sidecar  \
		--network=container:fortio-app \
		-u 1337:1337 \
        -v ${PWD}/conf/sidecar:/var/lib/istio/envoy \
	 envoyproxy/envoy-alpine-dev \
	         -c /var/lib/istio/envoy/bootstrap.yaml \
             --base-id 4 \
             --log-level trace \
             --restart-epoch 0 \
             --drain-time-s 45 \
             --parent-shutdown-time-s 60 \
             --service-cluster fortio \
             --service-node sidecar~10.244.1.1~fortio-7d467cd559-qrsbv.fortio~fortio.svc.cluster.local \
             --max-obj-name-len 189 \
             --local-address-ip-version v4 \
             --allow-unknown-fields


# Run a sidecar as a process on the VM (no docker), using pilot-agent and istio envoy
fortio-sidecar-local-agent:
	    ISTIO_META_NAMESPACE=fortio \
	    ${BINDIR}/pilot-agent proxy sidecar \
            --domain simple-micro.svc.cluster.local \
            --configPath ${TOP}/out \
            --binaryPath ${BINDIR}/envoy \
            --templateFile ${TOP}/src/istio.io/istio/tools/packaging/common/envoy_bootstrap_v2.json \
            --serviceCluster echosrv.fortio \
            --drainDuration 45s --parentShutdownDuration 1m0s \
            --discoveryAddress localhost:15010 \
            --proxyLogLevel=debug \
            --proxyComponentLogLevel=misc:info \
            --connectTimeout 10s \
            --proxyAdminPort 15000 \
            --concurrency 2 \
            --controlPlaneAuthPolicy NONE \
            --statusPort 15020 \
            --applicationPorts 8080,8079,8088 \
            --controlPlaneBootstrap=false

# Run envoy (istio version) as a process, no pilot-agent
fortio-sidecar-local:
	    ISTIO_META_NAMESPACE=fortio \
	     "${BINDIR}/envoy" -c conf/sidecar/bootstrap_local.yaml \
                --base-id 4 \
                --log-level debug \
                --restart-epoch 0 \
                --drain-time-s 45 \
                --parent-shutdown-time-s 60 \
                --service-cluster istio-ingressgateway \
                --service-node sidecar~10.244.1.82~fortio1.fortio~fortio.svc.cluster.local \
                --max-obj-name-len 189 \
                --local-address-ip-version v4 \
                --allow-unknown-fields

fortio-debug:
	docker run -it --rm --network=container:fortio-app --cap-add=NET_ADMIN  --entrypoint /bin/bash ${HUB}/proxyv2:${TAG}

fortio-iptables:
	docker stop fortio-iptables || true
	docker run -it --rm --name=fortio-iptables \
 		--network=container:fortio-app \
 		--cap-add=NET_ADMIN \
 		-e ISTIO_INBOUND_PORTS="*" \
 	 ${HUB}/proxy_init:${TAG}

fortio-app:
	docker stop fortio-app || true
	docker rm fortio-app || true
	docker ${DOCKER_START} --name=fortio-app \
		--add-host istio-pilot.istio-system:${IP} \
        -p 127.0.0.1:12080:8080 \
		-p 127.0.0.1:12006:15006 \
		-p 127.0.0.1:12001:15001 \
		-p 127.0.0.1:12000:15000 \
 	 fortio/fortio:latest

sidecar-test:
	docker run -it --rm --name sidecar-test \
	 -e ISTIO_CP_AUTH=NONE \
	 -e ISTIO_PILOT_PORT=15010 \
	 --add-host istio-pilot.istio-system:${IP} \
	 --cap-add=NET_ADMIN costinm/sidecar-test

# https://storage.googleapis.com/istio-release/releases/1.3.0-rc.0/deb/istio-sidecar.deb

# Simpler build for the components we need for testing.
# For real use the official docker images should be used.
build:
	#GOPROXY=https://proxy.golang.org
	GO111MODULE=on go get -u istio.io/istio/pilot/cmd/pilot-discovery
	GO111MODULE=on go install istio.io/istio/security/tools/generate_cert

# Generate root CA and certs for core components and tests
# This can be used on the local machine directly, and can also be provisioned in remote clusters
# with a script or ansible (using kubectl create-secret).
# It is expected users will use their own CA and upload it.
cacerts: conf/ca/ca-cert.pem
	# Returns 255 for some reason..
	${GOPATH}/bin/generate_cert -ca -organization cluster.local \
		-out-cert conf/ca/ca-cert.pem -out-priv conf/ca/ca-key.pem || true

	${GOPATH}/bin/generate_cert -client -organization cluster.local \
        -out-cert ${CONF}/pilot/cert-chain.pem -out-priv ${CONF}/pilot/key.pem \
        -signer-cert ${CONF}/ca/ca-cert.pem -signer-priv ${CONF}/ca/ca-key.pem \
        -host spiffe://cluster.local/ns/istio-system/sa/istio-pilot-service-account

KUBECONFIG ?= ${HOME}/.kube/config

K3S_OPT?=-it --rm

k3s-start:
	$(MAKE) k3s K3S_OPT="-d --restart always"

EXTRA_TLS=--tls-san 10.1.10.1

# Run a local k8s on the VM, in a docker container. Use k3s, with persistent directories.
# Note that the OUT dir will have logs and volumes, very easy to grep or automate.
k3s:
    # k3s/server/manifests - will have manifests
	mkdir -p ${OUT}/k3s/kube ${OUT}/k3s/var ${OUT}/k3s/run ${OUT}/k3s/k3s
	docker stop k3s || true
	docker rm k3s || true
	docker run ${K3S_OPT} --name k3s \
	 -e K3S_KUBECONFIG_OUTPUT=/output/kubeconfig.yaml \
	 -e K3S_KUBECONFIG_MODE=666 \
	 -p 6443:6443 \
	 -p 6080:80 \
	 -v ${OUT}/k3s/kube:/output \
	 -v ${OUT}/k3s/var:/var/ \
	 -v ${OUT}/k3s/run:/run \
	 -v ${OUT}/k3s/k3s:/var/lib/rancher/k3s \
	 --privileged \
	  rancher/k3s:latest \
	 server --https-listen-port 6443 ${EXTRA_TLS}

	docker exec -it k3s chmod 666 /output/kubeconfig.yaml


k3s-shell:
	docker exec -it k3s /bin/sh

build-image:
	docker build -f tools/build_img/Dockerfile -t costinm/istiod-build:latest .

knative-cluster:
	kubectl apply --selector knative.dev/crd-install=true \
		--filename https://github.com/knative/serving/releases/download/v0.11.0/serving.yaml \
		--filename https://github.com/knative/eventing/releases/download/v0.11.0/release.yaml \
		--filename https://github.com/knative/serving/releases/download/v0.11.0/monitoring.yaml

knative:
	kubectl apply --filename https://github.com/knative/serving/releases/download/v0.11.0/serving.yaml \
		--filename https://github.com/knative/eventing/releases/download/v0.11.0/release.yaml \
		--filename https://github.com/knative/serving/releases/download/v0.11.0/monitoring.yaml

# Swaps the container image.
okteto:
	#go run github.com/okteto/okteto up
	/ws/istio-stable/bin/okteto up

CHARTS=./manifests/charts
ISTIO_CHARTS=../istio/manifests/charts
include manifests/Makefile


# Configure istio-gate with custom config
deploy/gw-istio-gate-cfg-sample:
	# Service and Gateway object - name matches namespace.
	helm upgrade --install -n istio-gate \
		istio-gate samples/charts/istio-gate

USERCHARTS=./samples/charts
include samples/Makefile

# Install ingress using old templates, to verify migration to new template
deploy/old-ingress:
	helm upgrade -n istio-system --install \
		istio-ingressgateway ${ISTIO_CHARTS}/gateways/istio-ingress \
		--set revision=prod


deploy/k8s-registry:
	helm -n kube-registry upgrade --install kube-registry samples/charts/docker-registry


install/apps:
	kubectl apply -k test/fortio
	kubectl apply -k test/certmanager


uninstall/crd:
	kubectl delete crd -l release=istio || true

#####################################################3
# Building
#export TAG=latest
export HUB2 ?= costinm
#export HUB=localhost:30500
#export BUILD_WITH_CONTAINER=0

# Create docker images - no push needed.
# Works with K3S, minikube (docker or remote docker)
docker/pilot:
	cd ${ISTIO_GO} && $(MAKE) docker.pilot  DOCKER_ALL_VARIANTS=default

docker/proxyv2:
	cd ${TOP}/src/istio.io/istio && $(MAKE) docker.proxyv2  DOCKER_ALL_VARIANTS=default

images:
	cd ${TOP}/src/istio.io/istio && $(MAKE) docker.pilot docker.proxyv2  DOCKER_ALL_VARIANTS=default

push/pilot:
	cd ${TOP}/src/istio.io/istio && $(MAKE) push.docker.pilot  DOCKER_ALL_VARIANTS=default

push/proxyv2:
	cd ${TOP}/src/istio.io/istio && $(MAKE) push.docker.proxyv2  DOCKER_ALL_VARIANTS=default

# Push using TAG - to registry running in k8s
push:
	cd ${TOP}/src/istio.io/istio && $(MAKE) docker.pilot push.docker.proxyv2  DOCKER_ALL_VARIANTS=default

# Retag the local registry images, push do dockerhub
push/up:
	docker tag ${HUB}/pilot:${TAG} ${HUB2}/pilot:${TAG}
	docker tag ${HUB}/proxyv2:${TAG} ${HUB2}/proxyv2:${TAG}
	docker push ${HUB2}/pilot:${TAG}
	docker push ${HUB2}/proxyv2:${TAG}


gen:
	cd ${TOP}/src/istio.io/istio && \
	BUILD_WITH_CONTAINER=1 $(MAKE) fmt gen

prepareLocal:
	cd ${TOP}/src/istio.io/istio && \
		mkdir -p etc/certs && \
		cp -a  tests/testdata/certs/default/* etc/certs/

fetchCerts:
	cd ${TOP}/src/istio.io/istio/etc/certs && \
	go run istio.io/istio/security/tools/generate_cert


grpcurl-get:
	echo ${GOPATH}
	go get github.com/fullstorydev/grpcurl/cmd/grpcurl

events-watch:
	mkfifo /tmp/istiod1 || true
	sleep 100 > /tmp/istiod1 &
	cat /tmp/istiod1 | grpcurl -insecure localhost:15012 envoy.service.discovery.v2.AggregatedDiscoveryService/StreamAggregatedResources &
	sleep 1

	echo '{"node": {"id": "sidecar~1.1.1.1~debug~cluster.local", "metadata": {"GENERATOR": "event"}},"typeUrl": "istio.io/connections"}' > /tmp/istiod1

# Special makefile to build the image expected by skaffold
#
# Env:
#   IMAGES=costinm/pilot:TAG
#   HUB
#   PUSH_IMAGE=true
skaffold.istiod:
	docker tag costinm/pilot:latest ${IMAGE}

include samples/docker/Makefile

