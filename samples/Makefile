USERCHARTS?=./charts

# DNS mappings:
# ${GWDOMAIN} A ${GATEWAY_SVC_IP}
# (additional IPs possible - for example for migrations, hostport)
# *.${DOMAIN} CNAME ${GWDOMAIN}
#
# or
#
# *.${DOMAIN} A ${GATEWAY_SVC_IP}
DOMAIN?=k.webinf.info
SUBDOMAIN?=svc.i.webinf.info

deploy/samples: deploy/fortio deploy/httpbin deploy/iperf3 deploy/gw-cfg-istio-system

debug/configz:
	curl -H "Authorization: Bearer $(cat ./var/run/secrets/kubernetes.io/serviceaccount/token)"  istiod-debug.svc.i.webinf.info:8002/debug/configz

# Sample configuration for istio-ingressgateway.
#
deploy/gw-cfg-istio-system:
	helm upgrade --install \
		-n istio-system \
		gw-config \
 		${USERCHARTS}/cfg-ingressgateway \
 		--set domain=${DOMAIN} \
 		--set wildcardDomain=${SUBDOMAIN} \
 		-f ${USERCHARTS}/cfg-ingressgateway/values-costin.yaml


deploy/fortio:
	helm upgrade --install \
		-n fortio \
		fortio \
 		${USERCHARTS}/fortio

deploy/httpbin:
	helm upgrade --install \
		-n httpbin \
		httpbin \
 		${USERCHARTS}/httpbin

deploy/iperf3:
	helm upgrade --install \
		-n fortio \
		iperf3 \
 		${USERCHARTS}/iperf3
