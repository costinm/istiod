# Docker registry running in K8S, with Istio sidecar

Running a docker registry is easy - the image and persistent 
volume are sufficient. The harder problem is securing it - you
need certificates and DNS domain, and authentication is complicated
too.

This chart installs the registry as 'insecure' - but adds an
istio sidecar, with strict mtls mode. 

This allows access to be controlled by Istio policies - however
it doesn't solve the problem of how to use it. 

From a developer desktop you can port forward, and use localhost:5000
as docker registry - this is the special address that doesn't require
https. To test: curl localhost:5000/v2/

# Node access

To access the node we can deploy a Istio gateway of type 'host network' 
and as a DaemonSet.

The gateway listens on host port 5000, so kubelet has access. 


# Protecting port 5000


gcloud compute --project=${PROJECT} firewall-rules create \
  docker-local-registry \
  --description="block external access to port 5000 on node" \
  --direction=INGRESS --priority=1000 \
  --network=default --action=DENY \
  --rules=tcp:5000 --source-ranges=0.0.0.0/0

Test with: 

curl -v 35.225.53.10:5000/v2/ -HHost:localhost:5000
