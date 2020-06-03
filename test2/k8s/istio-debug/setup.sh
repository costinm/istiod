#!/usr/bin/env bash

# Domain - DNS entry must be created and point to the namespace of the ingress.
D=${1:-control.istio.webinf.info}

# Namespace of the gateway.
INGRESS=${2:-istio-ingress}

export ISTIOIP=$(kubectl get -n $INGRESS service ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

echo "Ingress IP: $ISTIOIP, DOMAIN: $D"

kubectl create ns istio-debug 2>/dev/null

cat test/k8s/istio-debug/pilot-gateway.yaml | \
  sed s/control.istio.webinf.info/$D/  | \
  sed s,istio-ingress,${INGRESS}, | \
  kubectl apply --grace-period=4  -f -
