#!/usr/bin/env bash

set -e



#RESOURCES="configmap secret daemonset deployment service hpa"

function kdumpContext() {
  local CONTEXT=$1

  local NAMESPACES=$(kubectl --context ${CONTEXT} get ns -o jsonpath="{.items[*].metadata.name}")
  local RESOURCES=$(kubectl --context ${CONTEXT} api-resources --namespaced -o name | tr "\n" " ")

      dir="${CONTEXT}/"
      mkdir -p "${dir}"
  kubectl --context ${CONTEXT} get crds -o yaml > ${CONTEXT}/crds.yaml

  #for ns in ${NAMESPACES};do
    for resource in ${RESOURCES};do
      #rsrcs=$(kubectl --context ${CONTEXT} -A get -o json ${resource}|jq '.items[].metadata.name'|sed "s/\"//g")
      #for r in ${rsrcs};do
      #  dir="${CONTEXT}/${ns}/${resource}"
      #  mkdir -p "${dir}"
      #  kubectl --context ${CONTEXT} -n ${ns} get -o yaml ${resource} ${r} > "${dir}/${r}.yaml"
      #done
      echo kubectl --context ${CONTEXT} get -A -o yaml ${resource}
      kubectl --context ${CONTEXT} get -A -o yaml ${resource} > "${dir}/${resource}.yaml" || true
    done
  #done

}

function kdumpAll() {
  CONTEXTS=$(kubectl config get-contexts -o name)

  for c in ${CONTEXTS}; do
    time kdumpContext $c
  done
}

kdumpAll
