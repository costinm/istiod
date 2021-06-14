# Install Istiod using the simplified charts.

# Revision tag for the istiod install. Mapped to canary.
REV?=v1-11

# Base directory for the charts.
CHARTS?=./charts
ISTIO_CHARTS?=../../istio/manifests/charts

helm/all: deploy/namespaces deploy/gate-lb-service \
		  deploy/istiod deploy/webhooks-canary \
		  deploy/gw-istio-system \
		  deploy/gw-istio-gate \
		  deploy/gw-hostport

deploy/namespaces:
	@kubectl create ns istio-system || true
	@kubectl create ns istio-gate || true
	@kubectl create ns istio-gw-hostport || true

# A single version of Istiod - using a version-based revision name.
# The version will be associated with labels using in the other charts.
deploy/istiod:
	# Install istiod.
	# Telemetry configs can be installed as a separate chart - this
	# avoids upgrade issues for 1.4 skip-version.
	# TODO: add telementry to docker image
	helm upgrade --install \
 		-n istio-system \
 		istiod-${REV} \
        ${ISTIO_CHARTS}/istio-control/istio-discovery \
		--set revision=${REV} \
		--set telemetry.enabled=true \
		--set meshConfig.accessLogFile=/dev/stdout \
		--set values.pilot.env.PILOT_ENABLE_WORKLOAD_ENTRY_AUTOREGISTRATION=true \
		--set values.pilot.env.PILOT_ENABLE_WORKLOAD_ENTRY_HEALTHCHECKS=true

deploy/webhooks-canary:
	helm upgrade --install \
        -n istio-system \
        istio-webhook-canary \
        ${CHARTS}/istio-webhook-tag \
		--set revision=${REV} \
		--set tag=canary

# Make this revision of istiod the default.
deploy/webhooks-default:
	helm -n istio-system upgrade --install istio-webhook-default manifests/charts/istio-webhook-tag \
		--set revision=${REV} --set enableIstioInjection=true

# Deploy a 'canary' gateway deployment in istio-system.
# This is selected by the gateway service and receives traffic, but
# does not overlap with the default gateway install.
#
# If migrating, make sure to 'take over' the Service before deleting
# the default gateway install.
deploy/gw-istio-system:
	helm upgrade --install \
		-n istio-system \
		gate-canary \
		${CHARTS}/gate \
		--set revision=canary \
		--set autoscaleMax=1
	# Default config - just 80 and 443


# Configure gateway service objects - allocates 2 LB IPs for
# the istio-system and istio-gate namespaces. The hostport gateway
# does not use a LB but node IPs.
deploy/gate-lb-service:
	# Service name matches namespace - it is best to keep each LB and Service
    # separated.
	kubectl apply -f ${CHARTS}/../gateway-class.yaml
	helm upgrade --install \
        -n istio-gate \
		istio-gate \
		${CHARTS}/gateway-service
	helm upgrade --install \
		-n istio-system \
		istio-ingressgateway \
		${CHARTS}/gateway-service \
		--set selector=ingressgateway

# 2 deployments, one using default and one using canary
# In the istio-gate namespace
deploy/gw-istio-gate:
    # No revision set - will use default injection.
    # Name of the install based on the revision - can be any string
	helm upgrade --install \
          -n istio-gate \
		  gate-default \
		  ${CHARTS}/gate \
		  --set autoscaleMax=1 \
		  --set routerMode=sni-dnat

	helm  upgrade --install \
            -n istio-gate \
			gate-canary \
			${CHARTS}/gate \
   		    --set autoscaleMax=1 \
     		--set routerMode=sni-dnat \
    		--set revision=canary

# Deploy a hostport-based gateway in istio-gw-hostport namespace.
# Will use 'canary' Istiod floating label
deploy/gw-hostport:
	helm upgrade --install \
         -n istio-gw-hostport \
         istio-gw-hostport \
		${CHARTS}/gate-hostport \
		--set routerMode=sni-dnat --set revision=canary

