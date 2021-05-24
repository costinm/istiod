
test: build canary/testca canary/testoss


managed/test:
	kubectl delete mutatingwebhookconfiguration istiod-asmca || true
	curl     --request POST     --header 'X-Server-Timeout: 600'     \
		--header "Authorization: Bearer $(shell gcloud auth print-access-token)"    \
		--header "Content-Type: application/json" \
        --data '{"image": "gcr.io/wlhe-cr/cloudrun:asm-canary"}' \
		 https://staging-meshconfig.sandbox.googleapis.com/v1alpha1/projects/${PROJECT_ID}/locations/${ZONE}/clusters/${CLUSTER}:runIstiod
	kubectl -n fortio-asmca rollout restart deployment
	kubectl wait deployments fortio -n fortio-asmca --for=condition=available --timeout=10s


asm-cloudrun/gvisor/test:
	PROJECT_ID=asm-cloudrun CLUSTER=gvisor TAG=-asmca REV=asmca $(MAKE) fortio


canary/testca:
	TAG=asm-canary PROJECT_ID=costin-asm1 CLUSTER=big1 REV=asmca-canary ENVEXTRA="ASM=1,CA=1," NS=fortio-asmca-canary $(MAKE) push_and_test

canary/testoss:
	TAG=asm-canary PROJECT_ID=costin-asm1 CLUSTER=big1 REV=oss-canary ENVEXTRA="" NS=fortio-oss-canary $(MAKE) push_and_test

costin-asm1/test3:
	gcloud container clusters get-credentials test3 --zone us-central1-c --project costin-asm1


ADDON_TAG=asm-addon

# K8S 1.18.10
# Istio 1.4.10
asm-cloudrun/addon:
	$(MAKE) test/addon PROJECT_ID=asm-cloudrun CLUSTER=addon ZONE=us-central1-c

# K8S 1.15, no WI
# Istio 1.4.10
asm-cloudrun/skip15:
	$(MAKE) test/addon PROJECT_ID=asm-cloudrun CLUSTER=skip15 ZONE=us-central1-c

# K8S 1.17.12
# Istio 1.4.10
asm-cloudrun/gvisor:
	$(MAKE) test/addon PROJECT_ID=asm-cloudrun CLUSTER=gvisor ZONE=us-central1-c

# NOT SUPPORTED: No 'objectselector'
# K8S 1.14.10
# Istio 1.2.10
asm-cloudrun/addon-14-nowi:
	$(MAKE) test/addon PROJECT_ID=asm-cloudrun CLUSTER=addon-14-nowi ZONE=us-central1-c


costin-asm1/big1/fortio-asm:
	PROJECT_ID=costin-asm1 CLUSTER=big1 TAG=asm-cr $(MAKE) run3
	PROJECT_ID=costin-asm1 CLUSTER=big1 TAG=asm-cr $(MAKE) fortio3
	PROJECT_ID=costin-asm1 CLUSTER=big1 $(MAKE) fortio/restart

costin-asm1/run:
	PROJECT_ID=costin-asm1 CLUSTER=big1 TAG=asm-cr $(MAKE) run3
	PROJECT_ID=costin-asm1 CLUSTER=cloudrun TAG=asm-cr $(MAKE) run3
	PROJECT_ID=costin-asm1 CLUSTER=test1 TAG=asm-cr $(MAKE) run3

costin-asm1/test:
	PROJECT_ID=costin-asm1 CLUSTER=big1 TAG=asm-cr $(MAKE) fortio3
	PROJECT_ID=costin-asm1 CLUSTER=cloudrun TAG=asm-cr $(MAKE) fortio3
	PROJECT_ID=costin-asm1 CLUSTER=test1 TAG=asm-cr $(MAKE) fortio3

costin-asm1/restart:
	PROJECT_ID=costin-asm1 CLUSTER=big1 $(MAKE) fortio/restart
	PROJECT_ID=costin-asm1 CLUSTER=cloudrun $(MAKE) fortio/restart
	PROJECT_ID=costin-asm1 CLUSTER=test1 $(MAKE) fortio/restart
	PROJECT_ID=costin-asm1 CLUSTER=big1 NS=test-ns $(MAKE) _restart



canary/build2:
	(cd ${ISTIO_GO} ; TAG=asm-canary REV=asm-canary make push.docker.cloudrun push.docker.proxyv2)

canary/deploy:
	TAG=asm-canary REV=asm-canary ENVEXTRA="ASM=1," $(MAKE) _run
	TAG=asm-canary REV=oss-canary ENVEXTRA="" $(MAKE) _run
	TAG=asm-canary REV=asmca-canary ENVEXTRA="ASM=1,CA=1," $(MAKE) _run

canary: canary/build canary/test

canary/build:
	(cd ${ISTIO_GO} ; TAG=asm-canary REV=asm-canary make push.docker.cloudrun )

canary/test:
	TAG=asm-canary PROJECT_ID=costin-asm1 CLUSTER=big1 REV=asm-canary $(MAKE) _run
	TAG=asm-canary PROJECT_ID=costin-asm1 CLUSTER=big1 REV=asm-canary $(MAKE) fortio
	TAG=asm-canary PROJECT_ID=costin-asm1 CLUSTER=big1 REV=asm-canary NS=fortio-asm-canary $(MAKE) _restart


asm-cloudrun/test:
	PROJECT_ID=asm-cloudrun CLUSTER=small TAG=asm-cr $(MAKE) fortio3
	PROJECT_ID=asm-cloudrun CLUSTER=addon TAG=asm-cr REV=asm $(MAKE) fortio3

asm-cloudrun/restart:
	PROJECT_ID=asm-cloudrun CLUSTER=small $(MAKE) fortio/restart
	PROJECT_ID=asm-cloudrun CLUSTER=addon $(MAKE) fortio/restart


asm-cloudrun/run:
	PROJECT_ID=asm-cloudrun CLUSTER=small $(MAKE) run3
	PROJECT_ID=asm-cloudrun CLUSTER=addon $(MAKE) run3

costin-demo1/run:
	PROJECT_ID=costin-demo1 CLUSTER=demo1 $(MAKE) run3

wlhe-cr/run:
	PROJECT_ID=wlhe-cr CLUSTER=istio $(MAKE) run3



CI_SA=prow-gob-storage@istio-prow-build.iam.gserviceaccount.com

fixci:

