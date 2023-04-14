BASE:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
TOOLS:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
ISTIOD:=$(shell dirname $(TOOLS))
TOP=$(shell cd ${ISTIOD}/../../..; pwd)
ISTIO_SRC=${TOP}/src/istio.io/istio
ISTIO_CHARTS?=${ISTIO_SRC}/manifests/charts/istio-control/istio-discovery

-include ${TOP}/src/istio.io/.local.mk

CHART_VERSION=--devel

HUB?=gcr.io/istio-testing
ISTIO_HUB?=${HUB}
export ISTIO_HUB
export HUB
ISTIO_TAG?=latest
export ISTIO_TAG

ISTIO_PROXY_IMAGE?=${ISTIO_HUB}/proxyv2:latest

echo: PROJECT_ID?=$(shell gcloud config get-value project)
echo:
	@echo BASE: ${BASE}
	@echo TOP: ${TOP}
	@echo MAKEFILE_LIST: $(MAKEFILE_LIST)
	@echo CLUSTER: ${CLUSTER_NAME}
	@echo PROJECT_ID: ${PROJECT_ID}
	@echo ISTIO_HUB: ${ISTIO_HUB}
	@echo TAG: ${TAG}


# Recommended method for GKE
#install-gateway-crd: PROJECT_ID=$(shell gcloud config get-value project)
install-gateway-crd:
    # gcloud components update may be required
	gcloud  container clusters update ${CLUSTER_NAME} \
		--gateway-api=standard \
		--project=${PROJECT_ID} \
		--region=${CLUSTER_LOCATION}

# Push using TAG - to registry running in k8s
push:
	echo ${TAG} ${HUB}
	cd ${TOP}/src/istio.io/istio && $(MAKE) push.docker.install-cni push.docker.ztunnel push.docker.pilot push.docker.proxyv2  DOCKER_ALL_VARIANTS=default

images:
	cd ${TOP}/src/istio.io/istio && $(MAKE) docker.pilot docker.proxyv2  DOCKER_ALL_VARIANTS=default

push/pilot:
	cd ${TOP}/src/istio.io/istio && $(MAKE) push.docker.pilot  DOCKER_ALL_VARIANTS=default

push/echoapp:
	cd ${TOP}/src/istio.io/istio && $(MAKE) push.docker.app  DOCKER_ALL_VARIANTS=default

push/proxyv2:
	cd ${TOP}/src/istio.io/istio && $(MAKE) push.docker.proxyv2  DOCKER_ALL_VARIANTS=default


helm/addcharts:
	helm repo add istio https://istio-release.storage.googleapis.com/charts
	helm repo update

deploy/istio-base:
	kubectl create namespace istio-system | true
	helm upgrade --install istio-base istio/base -n istio-system ${CHART_VERSION} | true

# Istio-system doesn't work on GKE - with priority-class-critical

deploy/cni:
	(cd ${ISTIO_SRC}/manifests/charts; helm upgrade --install  \
          -n kube-system \
           istio-cni \
            istio-cni \
            --set global.hub=${ISTIO_HUB} \
            --set global.tag=${ISTIO_TAG} \
        --set cni.logLevel=info --set cni.privileged=true --set cni.ambient.enabled=true \
        --set cni.excludeNamespaces[0]=kube-system \
        ${CNI_EXTRA} )

deploy/cni-t:
	(cd ${ISTIO_SRC}/manifests/charts; helm template  \
          -n kube-system \
           \
            istio-cni \
            --set global.hub=${ISTIO_HUB} \
            --set global.tag=${ISTIO_TAG} \
            --set cni.cniBinDir=/home/kubernetes/bin \
        --set cni.logLevel=info \
        --set cni.privileged=true \
        --set cni.ambient.enabled=true \
        --set cni.ambient.redirectMode=ebpf \
        ${CNI_EXTRA}  | kubectl apply -f -)

# --set cni.ambient.redirectMode=ebpf for calico

deploy/ztunnel:
	(cd ${ISTIO_SRC}/manifests/charts; helm template  \
          -n istio-system \
           \
            ztunnel \
            --set global.hub=${ISTIO_HUB} \
            --set global.tag=${ISTIO_TAG} \
        ${ZTUNNEL_EXTRA} | kubectl apply -f -)


template/ztunnel:
	(cd ${ISTIO_SRC}/manifests/charts; helm template  \
          -n istio-system \
           \
            ztunnel \
            --set global.hub=${ISTIO_HUB} \
            --set global.tag=${ISTIO_TAG} \
            --set redirectMode=ebpf \
        ${ZTUNNEL_EXTRA})

install/helm:
	curl -fsSL -o helm.tgz https://get.helm.sh/helm-v3.11.1-linux-amd64.tar.gz
	tar xvfz helm.tgz
	mv linux-amd64/helm ${OUT}/helm
	rm -rf helm.tgz linux-amd64

helm/takeover:
	kubectl annotate clusterrole istio-cni meta.helm.sh/release-name=istio-cni meta.helm.sh/release-namespace=istio-system --overwrite
	kubectl label clusterrole istio-cni app.kubernetes.io/managed-by=Helm
	kubectl annotate clusterrole istio-cni-repair-role meta.helm.sh/release-name=istio-cni meta.helm.sh/release-namespace=istio-system --overwrite
	kubectl label clusterrole istio-cni-repair-role app.kubernetes.io/managed-by=Helm

# Default install of istiod, with a number of options set for interop with ASM and MCP.
#
# TODO: add docs on how to upgrade an existing istio, explain the config.
#
# To install a revisioned istio, replace "istiod" with "istiod-REV and add --set revision=${REV}
#
# Note that trustDomain is set to the value used by ASM - on GKE this is important since it allows getting access
# tokens. If using istio-ca ( standard istio ), OSS_ISTIO=true must be set when starting the app, to get the right
# type of token. TODO: trust domain should be included in the mesh-env and used from there.
#deploy/istiod: CONFIG_PROJECT_ID?=$(shell gcloud config get-value project)

deploy/istiod:
	$(MAKE) _deploy/istiod

deploy/istiod117:
	$(MAKE) _deploy/istiod ISTIO_CHARTS=

_deploy/istiod:
	#helm template --debug
	helm upgrade --install \
       	-n istio-system \
 		istiod \
        ${ISTIO_CHARTS} \
        ${CHART_VERSION} \
        ${ISTIOD_EXTRA} \
        --set global.hub=${ISTIO_HUB} \
        --set global.tag=${ISTIO_TAG} \
        --set global.imagePullPolicy=Always \
       	--set pilot.resources.requests.cpu=100m \
       	--set pilot.resources.requests.memory=512Mi \
        --set pilot.replicaCount=1 \
        --set pilot.autoscaleEnabled=false \
         \
		--set telemetry.enabled=false \
		--set telemetry.v2.enabled=false \
		\
        --set meshConfig.proxyHttpPort=15007 \
        \
        --set meshConfig.accessLogFile=/dev/stdout \
        --set meshConfig.defaultProviders.metrics[0]=prometheus \
        --set meshConfig.extensionProviders[0].name=prometheus \
        --set meshConfig.extensionProviders[0].prometheus={} \
        \
        --set pilot.podAnnotations.'security\.cloud\.google\.com/use-workload-certificates'="" \
        \
        --set pilot.env.VERIFY_CERTIFICATE_AT_CLIENT=true \
        --set pilot.env.PILOT_VERIFY_CERTIFICATE_AT_CLIENT=true \
           \
        --set meshConfig.defaultConfig.proxyMetadata.ISTIO_META_ENABLE_HBONE="true" \
        --set pilot.env.PILOT_ENABLE_AMBIENT_CONTROLLERS=true \
        --set pilot.env.PILOT_ENABLE_HBONE=true \
        \
 		--set meshConfig.trustDomain="cluster.local" \
        --set meshConfig.trustDomainAliases[0]="${CONFIG_PROJECT_ID}.svc.id.goog" \
        --set global.sds.token.aud="${CONFIG_PROJECT_ID}.svc.id.goog" \
        --set pilot.env.CA_TRUSTED_NODE_ACCOUNTS="istio-system/ztunnel\,kube-system/ztunnel" \
		--set pilot.env.TOKEN_AUDIENCES="${CONFIG_PROJECT_ID}.svc.id.goog\,istio-ca" \
        --set pilot.env.ISTIO_MULTIROOT_MESH="true" \
        \
        --set pilot.env.PILOT_ENABLE_WORKLOAD_ENTRY_AUTOREGISTRATION=true \
		--set pilot.env.PILOT_ENABLE_WORKLOAD_ENTRY_HEALTHCHECKS=true \
		\
        --set pilot.env.PILOT_ENABLE_PERSISTENT_SESSION_FILTER=true \
        \
		--set pilot.env.ENABLE_MCS_HOST=true \
		--set pilot.env.ENABLE_MCS_SERVICE_DISCOVERY=true \
		--set pilot.env.MCS_API_GROUP=net.gke.io \
		--set pilot.env.MCS_API_VERSION=v1

#        --set pilot.env.PILOT_ENABLE_INBOUND_PASSTHROUGH=false \

# Multi-network via auto-sni - replaced with hbone
#        --set pilot.env.ENABLE_AUTO_SNI=true \

# MCS

# Ambient:
# 	  PILOT_ENABLE_INBOUND_PASSTHROUGH=true \

deploy/addons:
	kubectl apply -f https://raw.githubusercontent.com/istio/istio/master/samples/addons/prometheus.yaml
	kubectl apply -f https://raw.githubusercontent.com/istio/istio/master/samples/addons/kiali.yaml


deploy/ambient-stability:
	kubectl apply -f ${ISTIOD}/manifests/charts/ambient-stability/namespace.yaml
	helm upgrade --install \
       	-n ambient-stability ambient-stability \
        ${ISTIOD}/manifests/charts/ambient-stability


deploy/iperf-ztunnel:
	kubectl apply -f ${ISTIOD}/manifests/charts/ambient-stability/namespace.yaml
	helm upgrade --install --set waypoint=namespace \
       	-n iperf-ztunnel iperf-ztunnel \
        ${ISTIOD}/manifests/charts/ambient-stability


tcpdump:
	#kubectl -n ${NS} patch pod ${POD} --patch $(shell cat tcpdump_patch.yaml)
	#kubectl alpha debug -it ${POD} --image=busybox --target=mando -- /bin/sh
	#kubectl exec ${POD} -c nginx -- tcpdump -i eth0 -w - | wireshark -k -i -
	kubectl alpha -n ${NS} debug -i ${POD} --image=nicolaka/netshoot --target=tcpdump -- tcpdump -i eth0 -w - | wireshark -k -i -

poddebug:
	#kubectl -n ${NS} patch pod ${POD} --patch $(shell cat tcpdump_patch.yaml)
	#kubectl alpha debug -it ${POD} --image=busybox --target=mando -- /bin/sh
	#kubectl exec ${POD} -c nginx -- tcpdump -i eth0 -w - | wireshark -k -i -
	kubectl alpha -n ${NS} debug -i ${POD} --image=nicolaka/netshoot --target=debug -- /bin/bash

hostdebug:
	kubectl run tmp-shell --rm -it \
		--overrides='{"spec": {"hostNetwork": true}}' \
		--image nicolaka/netshoot -- /bin/bash

deploy/argo:
	kubectl create namespace argo-rollouts || true
	kubectl apply -n argo-rollouts -f https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml
#	kubectl create clusterrolebinding \
#		YOURNAME-cluster-admin-binding \
#		--clusterrole=cluster-admin \
#		--user=YOUREMAIL@gmail.com
	# docker run quay.io/argoproj/kubectl-argo-rollouts:master version

secret:
	istioctl pc secret ds/ztunnel -n istio-system -o json | jq -r '.dynamicActiveSecrets[0].secret.tlsCertificate.certificateChain.inlineBytes' | base64 --decode | openssl x509 -noout -text -in /dev/stdin

