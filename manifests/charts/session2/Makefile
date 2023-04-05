include ../../tools/common.mk

debug: echo
	skaffold dev --cleanup=false --namespace=session

install:
	kubectl get crd gateways.gateway.networking.k8s.io || \
      { kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v0.5.1" | kubectl apply -f -; }


mkcluster:
	gcloud container --project "costin-asm1" clusters create-auto "no-asm" --region "us-central1" \
	--release-channel "rapid" --network "projects/costin-asm1/global/networks/default" \
	 --subnetwork "projects/costin-asm1/regions/us-central1/subnetworks/default" \
	 --cluster-ipv4-cidr "/17" --services-ipv4-cidr "/22"

dump: POD=$(shell kubectl -n session get pod -l istio=session -o jsonpath='{.items[0].metadata.name}')
dump:
	istioctl pc l ${POD}.session --port 8080 -o yaml
