BASE?=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
TOOLS?=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
TOP=$(shell cd ${TOOLS}/..; pwd)

-include ${TOP}/.local.mk

PROJECT_ID?=$(shell gcloud config get-value project)
CLUSTER_NAME?=big1
CLUSTER_LOCATION?=us-central1-c

ISTIO_CHARTS?=istio/istiod
CHART_VERSION=--devel

ISTIO_HUB?=gcr.io/istio-testing
export ISTIO_HUB
ISTIO_TAG?=latest
export ISTIO_TAG

ISTIO_PROXY_IMAGE?=${ISTIO_HUB}/proxyv2:latest

echo:
	@echo BASE: ${BASE}
	@echo TOP: ${TOP}
	@echo CLUSTER: ${CLUSTER_NAME}
	@echo PROJECT_ID: ${PROJECT_ID}
	@echo ISTIO_HUB: ${ISTIO_HUB}

# Recommended method for GKE
install-gateway-crd:
    # gcloud components update may be required
	gcloud  container clusters update ${CLUSTER_NAME} \
		--gateway-api=standard \
		--project=${PROJECT_ID} \
		--region=${CLUSTER_LOCATION}



helm/addcharts:
	helm repo add istio https://istio-release.storage.googleapis.com/charts
	helm repo update

deploy/istio-base:
	kubectl create namespace istio-system | true
	helm upgrade --install istio-base istio/base -n istio-system ${CHART_VERSION} | true

# Default install of istiod, with a number of options set for interop with ASM and MCP.
#
# TODO: add docs on how to upgrade an existing istio, explain the config.
#
# To install a revisioned istio, replace "istiod" with "istiod-REV and add --set revision=${REV}
#
# Note that trustDomain is set to the value used by ASM - on GKE this is important since it allows getting access
# tokens. If using istio-ca ( standard istio ), OSS_ISTIO=true must be set when starting the app, to get the right
# type of token. TODO: trust domain should be included in the mesh-env and used from there.
deploy/istiod:
	helm upgrade --install \
 		-n istio-system \
 		istiod \
        ${ISTIO_CHARTS} \
        ${CHART_VERSION} \
        ${ISTIOD_EXTRA} \
        --set global.hub=${ISTIO_HUB} \
        --set global.tag=${ISTIO_TAG} \
        --set global.imagePullPolicy=Always \
		--set telemetry.enabled=true \
		--set global.sds.token.aud="${CONFIG_PROJECT_ID}.svc.id.goog" \
        --set meshConfig.trustDomain="cluster.local" \
        --set meshConfig.trustDomainAliases[0]="${CONFIG_PROJECT_ID}.svc.id.goog" \
             \
		--set meshConfig.proxyHttpPort=15007 \
        --set meshConfig.accessLogFile=/dev/stdout \
        \
        --set pilot.replicaCount=1 \
        --set pilot.autoscaleEnabled=false \
        --set pilot.podAnnotations.'security\.cloud\.google\.com/use-workload-certificates'="" \
        \
		--set pilot.env.TOKEN_AUDIENCES="${CONFIG_PROJECT_ID}.svc.id.goog\,istio-ca" \
        --set pilot.env.ISTIO_MULTIROOT_MESH=true \
        --set pilot.env.PILOT_ENABLE_WORKLOAD_ENTRY_AUTOREGISTRATION=true \
		--set pilot.env.PILOT_ENABLE_WORKLOAD_ENTRY_HEALTHCHECKS=true

