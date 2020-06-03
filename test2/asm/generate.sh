#!/bin/bash


#gcloud auth login
#gcloud config set project ${PROJECT_ID}

PROJECT_ID=`gcloud config list --format 'value(core.project)' 2>/dev/null`

#gcloud config set compute/zone ${ZONE}


# gcloud compute firewall-rules create csm-allow-health-checks --allow=tcp --source-ranges=130.211.0.0/22,35.191.0.0/16


NETWORK_NAME=$(basename $(gcloud container clusters describe $CLUSTER --project $PROJECT_ID --zone=$ZONE \
    --format='value(networkConfig.network)'))
SUBNETWORK_NAME=$(basename $(gcloud container clusters describe $CLUSTER --project $PROJECT_ID \
    --zone=$ZONE --format='value(networkConfig.subnetwork)'))

# Getting network tags is painful. Get the instance groups, map to an instance,
# and get the node tag from it (they should be the same across all nodes -- we don't
# know how to handle it, otherwise).
INSTANCE_GROUP=$(gcloud container clusters describe $CLUSTER --project $PROJECT_ID --zone=$ZONE --format='flattened(nodePools[].instanceGroupUrls[].scope().segment())' |  cut -d ':' -f2 | tr -d [:space:])
INSTANCE=$(gcloud compute instance-groups list-instances $INSTANCE_GROUP --project $PROJECT_ID \
    --zone=$ZONE --format="value(instance)" --limit 1)
NETWORK_TAGS=$(gcloud compute instances describe $INSTANCE --zone=$ZONE --project $PROJECT_ID --format="value(tags.items)")


cat <<EOF > neg/configmap-neg.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gce-config
  namespace: kube-system
  labels:
    release: asm
data:
  gce.conf: |
    [global]
    token-url = nil
    # Your cluster's project
    project-id = ${PROJECT_ID}
    # Your cluster's network
    network-name =  ${NETWORK_NAME}
    # Your cluster's subnetwork
    subnetwork-name = ${SUBNETWORK_NAME}
    # Prefix for your cluster's IG
    node-instance-prefix = gke-${CLUSTER}
    # Network tags for your cluster's IG
    node-tags = ${NETWORK_TAGS}
    # Zone the cluster lives in
    local-zone = ${ZONE}
EOF



cat <<EOF > configmap-galley.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: istiod-galley
  namespace: istio-system
  labels:
    release: asm
data:
  ISTIOD_ADDR: istiod-asm.istio-system.svc:15012
EOF

cat <<EOF > configmap-galley-master.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: istiod-galley
  namespace: istio-system
  labels:
    release: asm
data:
  galley.json: |
      {
      "EnableServiceDiscovery": true,
      "SinkAddress": "meshconfig.googleapis.com:443",
      "SinkAuthMode": "GOOGLE",
      "ExcludedResourceKinds": ["Pod", "Node", "Endpoints"],
      "SinkMeta": ["project_id=${PROJECT_ID}"]
      }

  PROJECT_ID: ${PROJECT_ID}
  GOOGLE_APPLICATION_CREDENTIALS: /var/secrets/google/key.json
  ISTIOD_ADDR: istiod-asm.istio-system.svc:15012


EOF
