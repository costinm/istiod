
fortio/restart:
	kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} -n fortio-asmca delete rs --all
	kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} -n fortio-oss delete rs --all

_restart:
	#kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} -n ${NS} delete rs --all
	kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} -n ${NS} rollout restart deployment
	kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} wait deployments fortio -n ${NS} --for=condition=available --timeout=10s


_delete:
	kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} -n ${NS} delete rs --all

# Run fortio in fortio-$REV, with $REV istio env
fortio:
	cat test/fortio.yaml | envsubst | kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} apply -f -

# Start all versions of fortio
fortio3:
	REV=asm-managed  $(MAKE) fortio
	REV=asmca $(MAKE) fortio
	REV=ossmanaged $(MAKE) fortio

fortioasm:
	REV=asm-168-9 $(MAKE) fortio

## TAG=asm-canary PROJECT_ID=costin-asm1 CLUSTER=big1 REV=asmca-canary ENVEXTRA="ASM=1,CA=1," NS=fortio-asmca-canary
push_and_test:
	$(MAKE) _run
	$(MAKE) fortio
	$(MAKE) _restart

test/addon: KUBECTL=kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER}
test/addon:
	gcloud container clusters get-credentials ${CLUSTER} --zone ${ZONE} --project ${PROJECT_ID}
	curl     --request POST     --header 'X-Server-Timeout: 600'     \
		--header "Authorization: Bearer $(shell gcloud auth print-access-token)"    \
		--header "Content-Type: application/json" \
        --data '{"image": "gcr.io/wlhe-cr/cloudrun:${ADDON_TAG}"}' \
		 https://staging-meshconfig.sandbox.googleapis.com/v1alpha1/projects/${PROJECT_ID}/locations/${ZONE}/clusters/${CLUSTER}:runIstiod
	${KUBECTL} create ns fortio-ossmanaged || true
	${KUBECTL} label ns fortio-ossmanaged istio.io/rev=ossmanaged || true
	# Verify mutating webhook exists
	${KUBECTL} get mutatingwebhookconfiguration istiod-ossmanaged

	REV=ossmanaged $(MAKE) fortio
	kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} -n fortio-ossmanaged delete rs --all
	istioctl x create-remote-secret --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} --name ${PROJECT_ID}/${ZONE}/${CLUSTER} > gke_${PROJECT_ID}_${ZONE}_${CLUSTER}.secret.yaml
