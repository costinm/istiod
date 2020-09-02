
fortio/restart:
	kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} -n fortio-asm delete rs --all
	kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} -n fortio-asmca delete rs --all
	kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} -n fortio-oss delete rs --all

_restart:
	kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} -n ${NS} delete rs --all

# Run fortio in fortio-$REV, with $REV istio env
fortio:
	cat test/fortio.yaml | envsubst | kubectl --context gke_${PROJECT_ID}_${ZONE}_${CLUSTER} apply -f -

fortio3:
	REV=asm $(MAKE) fortio
	REV=asmca TAG=-asmca $(MAKE) fortio
	REV=ossmanaged $(MAKE) fortio

fortioasm:
	REV=asm-168-9 $(MAKE) fortio
