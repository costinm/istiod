BASE?=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
TOOLS?=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
TOP=$(shell cd ${TOOLS}/..; pwd)

-include ../local.mk

PROJECT_ID?=$(shell gcloud config get-value project)
CLUSTER_NAME?=big1
CLUSTER_LOCATION?=us-central1-c

echo:
	@echo BASE: ${BASE}
	@echo TOP: ${TOP}
	@echo CLUSTER: ${CLUSTER_NAME}
	@echo PROJECT_ID: ${PROJECT_ID}

# Recommended method for GKE
install-gateway-crd:
    # gcloud components update may be required
	gcloud  container clusters update ${CLUSTER_NAME} \
		--gateway-api=standard \
		--project=${PROJECT_ID} \
		--region=${CLUSTER_LOCATION}
