#!/usr/bin/env bash

set -e



#RESOURCES="configmap secret daemonset deployment service hpa"

function kdumpContext() {
  local CONTEXT=$1

  local NAMESPACES=$(kubectl get ns -o jsonpath="{.items[*].metadata.name}")
  local RESOURCES=$(kubectl api-resources --namespaced -o name | tr "\n" " ")

  for ns in ${NAMESPACES};do
    for resource in ${RESOURCES};do
      rsrcs=$(kubectl --context ${CONTEXT} -n ${ns} get -o json ${resource}|jq '.items[].metadata.name'|sed "s/\"//g")
      for r in ${rsrcs};do
        dir="${CONTEXT}/${ns}/${resource}"
        mkdir -p "${dir}"
        kubectl --context ${CONTEXT} -n ${ns} get -o yaml ${resource} ${r} > "${dir}/${r}.yaml"
      done
    done
  done
}

function kdumpAll() {
  CONTEXTS=$(kubectl config get-contexts -o name)

  for c in ${CONTEXTS}; do
    kdumpContext $c
  done
}
