BASE=$(shell cd .; pwd)
TOP=$(shell cd ${BASE}/../../..; pwd)
GOPATH=${TOP}

CONF ?= ${BASE}/conf
HUB ?= gcr.io/istio-release
TAG ?= master-latest-daily

LOG_DIR ?= /tmp

pilot:
	docker stop pilot || true
	docker run -it --rm --name=pilot  \
		-p 127.0.0.1:12080:8080 \
		-p 127.0.0.1:12010:15010 \
		-p 127.0.0.1:12014:15014 \
		-p 127.0.0.1:129876:9876 \
        -v ${PWD}/conf/pilot:/var/lib/istio/pilot \
		-v ${PWD}/conf/istio:/var/lib/istio/istio \
	 ${HUB}/pilot:${TAG} \
    	 discovery --meshConfig /var/lib/istio/pilot/mesh.yaml \
    	--secureGrpcAddr="" \
    	--plugins="" \
        --configDir /var/lib/istio/istio --registries=MCP \
        --networksConfig /var/lib/istio/pilot/meshNetworks.yaml

start-local-pilot:
	kill -9 $(shell cat ${LOG_DIR}/pilot.pid) | true
	${GOPATH}/bin/pilot-discovery discovery \
        --meshConfig test/simple/mesh.yaml \
        --configDir ${CONF}/istio --registries=MCP \
        --networksConfig test/simple/meshNetworks.yaml & echo $$! > ${LOG_DIR}/pilot.pid

build:
	GO111MODULE=on go install istio.io/istio/pilot/cmd/pilot-discovery
	GO111MODULE=on go install istio.io/istio/security/tools/generate_cert

# Generate root CA and certs for core components and tests
# This can be used on the local machine directly, and can also be provisioned in remote clusters
# with a script or ansible (using kubectl create-secret).
# It is expected users will use their own CA and upload it.
cacerts:
	# Returns 255 for some reason..
	${GOPATH}/bin/generate_cert -ca -organization cluster.local \
		-out-cert ${CONF}/ca/ca-cert.pem -out-priv ${CONF}/ca/ca-key.pem || true

	${GOPATH}/bin/generate_cert -client -organization cluster.local \
        -out-cert ${CONF}/pilot/cert-chain.pem -out-priv ${CONF}/pilot/key.pem \
        -signer-cert ${CONF}/ca/ca-cert.pem -signer-priv ${CONF}/ca/ca-key.pem \
        -host spiffe://cluster.local/ns/istio-system/sa/istio-pilot-service-account


# Run citadel locally, in a docker container.
#
citadel:
	docker stop pilot || true
	docker run -it --rm --name=pilot  \
		-p 127.0.0.1:12080:8080 \
		-p 127.0.0.1:12010:15010 \
		-p 127.0.0.1:12014:15014 \
		-p 127.0.0.1:129876:9876 \
        -v ${PWD}/conf/pilot:/var/lib/istio/pilot \
		-v ${PWD}/conf/istio:/var/lib/istio/istio \
	 ${HUB}/pilot:${TAG} \
    	 discovery --meshConfig /var/lib/istio/pilot/mesh.yaml \
    	--secureGrpcAddr="" \
    	--plugins="" \
        --configDir /var/lib/istio/istio --registries=MCP \
        --networksConfig /var/lib/istio/pilot/meshNetworks.yaml
