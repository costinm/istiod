BASE=$(shell cd .; pwd)
TOP=$(shell cd ${BASE}/../../..; pwd)
GOPATH=${TOP}

CONF ?= ${BASE}/conf
HUB ?= gcr.io/istio-release
TAG ?= master-latest-daily

LOG_DIR ?= /tmp

# Namespace to use for the test app
NS=fortio

IP ?= $(shell hostname --ip-address)

# Start pilot in a docker container, using a local set of files, no k8s used.
# If configDir is specified, it will be used as a direct source of config, instead of CRDs. Will skip creation
# of the kubeClient as long as 'registries' doesn't include k8s.
#
# "--registries" still needed, to disable k8s registry - even if MCP is not configured in mesh.yaml
# ServiceEntries will be loaded from the config dir.
#
# Plugins: authz, authn, mixer, health
pilot:
	docker stop pilot || true
	docker run -it --rm --name=pilot  \
		-p 127.0.0.1:15080:8080 \
		-p 0.0.0.0:15010:15010 \
		-p 127.0.0.1:15014:15014 \
		-p 127.0.0.1:15876:9876 \
        -v ${PWD}/conf/pilot:/var/lib/istio/pilot \
		-v ${PWD}/conf/istio:/var/lib/istio/istio \
		-e PILOT_ENABLE_PROTOCOL_SNIFFING=true \
	 ${HUB}/pilot:${TAG} \
    	 discovery --meshConfig /var/lib/istio/pilot/mesh.yaml \
    	--secureGrpcAddr="" \
    	--plugins="authz" \
        --configDir /var/lib/istio/istio --registries=MCP \
        --networksConfig /var/lib/istio/pilot/meshNetworks.yaml


# A second Pilot instance, but using Galley config. Both pilots are based on the same config dir.
pilot-galley:
	yq m conf/pilot/mesh.yaml conf/pilot/mesh-galley.yaml > conf/pilot/gen-mesh-galley.yaml
	yq w -i conf/pilot/gen-mesh-galley.yaml configSources[0].address ${IP}:9901
	docker stop pilot-galley || true
	docker run -it --rm --name=pilot-galley  \
		-p 127.0.0.1:16080:8080 \
		-p 0.0.0.0:16010:15010 \
		-p 127.0.0.1:16014:15014 \
		-p 127.0.0.1:16876:9876 \
        -v ${PWD}/conf/pilot:/var/lib/istio/pilot \
		-v ${PWD}/conf/istio:/var/lib/istio/istio \
		-e PILOT_ENABLE_PROTOCOL_SNIFFING=true \
	 ${HUB}/pilot:${TAG} \
    	 discovery --meshConfig /var/lib/istio/pilot/gen-mesh-galley.yaml \
    	--secureGrpcAddr="" \
    	--plugins="authz" \
        --registries=MCP \
        --networksConfig /var/lib/istio/pilot/meshNetworks.yaml


# Start galley, using a local directory as config source.
# Passing kubeconfig instead of configPath will use K8S server, file must be included in the galley directory or mounted.
galley:
	docker stop galley || true
	docker run -it --rm --name=galley  \
		-p 0.0.0.0:9901:9901 \
		-p 127.0.0.1:15015:15015 \
		-p 127.0.0.1:15877:9877 \
        -v ${PWD}/conf/pilot:/var/lib/istio/pilot \
        -v ${PWD}/conf/galley:/var/lib/istio/galley \
		-v ${PWD}/conf/istio/test:/var/lib/istio/istio \
	 ${HUB}/galley:${TAG} \
    	 server -c /var/lib/istio/galley/galley.yaml \
    	    --meshConfigFile /var/lib/istio/pilot/mesh.yaml \
			--configPath /var/lib/istio/istio

# Same as pilot, but running on local machine. Easy to attach a debugger/step.
#
run-local-pilot:
	#kill -9 $(shell cat ${LOG_DIR}/pilot.pid) | true
	PILOT_ENABLE_PROTOCOL_SNIFFING=true \
	 ${GOPATH}/bin/pilot-discovery discovery \
        --meshConfig conf/pilot/mesh.yaml \
    	--plugins="authz" \
        --configDir conf/istio --registries=MCP \
        --networksConfig test/simple/meshNetworks.yaml # & echo $$! > ${LOG_DIR}/pilot.pid


# Start a local gateway, running in docker. Uses upstream envoy
# We need to pass the pilot address - either as an /etc/host set to host IP (or another addr),
# or as 127.0.0.1 and use the network of the pilot container.
gateway:
	docker stop gateway || true
	docker run -it --rm --name=gateway  \
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

# Set to "run -it --rm " for debugging
#DOCKERSTART ?= run -d
DOCKERSTART ?= run -it --rm

fortio-sidecar:
	docker stop fortio-sidecar || true
	docker rm fortio-sidecar || true
	docker ${DOCKERSTART} --name=fortio-sidecar  \
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

BINDIR=${TOP}/out/linux_amd64/release

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
	docker ${DOCKERSTART} --name=fortio-app \
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

gateway-shell:
	# apk add curl, ...
	docker exec -it gateway sh

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
	docker stop citadel || true
	docker run -it --rm --name=citadel  \
		-p 127.0.0.1:12080:8080 \
		-p 127.0.0.1:12010:15010 \
		-p 127.0.0.1:12014:15014 \
		-p 127.0.0.1:129876:9876 \
        -v ${PWD}/conf/pilot:/var/lib/istio/pilot \
		-v ${PWD}/conf/istio:/var/lib/istio/istio \
	 ${HUB}/citadel:${TAG} \
    	 discovery --meshConfig /var/lib/istio/pilot/mesh.yaml \
    	--secureGrpcAddr="" \
    	--plugins="" \
        --configDir /var/lib/istio/istio --registries=MCP \
        --networksConfig /var/lib/istio/pilot/meshNetworks.yaml

start-local-kind:
	kind start cluster --name local

