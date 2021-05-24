#!/usr/bin/env bash

kubectl create ns istio-micro-ingress 2>/dev/null
kubectl create ns test-micro-ingress 2>/dev/null

# --prune --all --cascade=true  -> doesn't work across namespaces
cat test/k8s/minimal-ingress/istio-micro-ingress/* \
  | sed s/example.com/${DOMAIN:-example.com}/ | kubectl apply --grace-period=4 -n istio-micro-ingress -f -
cat test/k8s/minimal-ingress/test-micro-ingress/* \
  | sed s/example.com/${DOMAIN:-example.com}/ | kubectl apply --grace-period=4 -n test-micro-ingress -f -

kubectl get nodes -o wide
