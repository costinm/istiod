REPO ?= costinm

docker/proxybase:
	docker build -t ${REPO}/proxybase:latest .

push/proxybase:
	docker push ${REPO}/proxybase:latest

docker/fortio:
	docker build -f Dockerfile.fortio ${REPO}/proxyfortio:latest .

push/fortio:
	docker push ${REPO}/proxyfortio:latest

