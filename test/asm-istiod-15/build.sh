
# All resources for istio-system, no cluster wide resource
helm template --name=asm-istio --namespace=istio-system ${ISTIO_SRC}/manifests/istio-control/istio-discovery \
  -f ${ISTIO_SRC}/manifests/global.yaml \
    	  --set clusterResources=false \
    	  --set revision=asm15 \
        --set global.useMCP=false > cluster/istio-1.5.yaml

# All cluster resources - roles, bindings, namespaces, service accounts.
# ServiceAccounts must be here because ClusterBindings depend on SA.
helm template --name=asm-istio --namespace=istio-system \
  ${ISTIO_SRC}/manifests/istio-control/istio-discovery  \
    	  -f ${ISTIO_SRC}/manifests/global.yaml \
    	  --set clusterResources=true \
    	  --set global.operatorManageWebhooks=true \
    		--set revision=asm15 > cluster/roles.yaml

cp ${ISTIO_SRC}/manifests/base/files/crd-all.gen.yaml cluster/

# Should works alongside istio OSS installed with:
# istioctl manifest apply  -s hub=gcr.io/istio-testing -s tag=latest
# and older versions
