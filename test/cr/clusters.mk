costin-asm1/big1:
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

canary: canary/build canary/test

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
