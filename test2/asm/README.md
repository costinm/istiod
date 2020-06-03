# Istio config with TD/Managed CA

We'll use 3 env variables that must be set first:

- PROJECT_ID
- CLUSTER
- ZONE

# Project setup 

## Enable firewall rules for health check

```shell script

gcloud compute firewall-rules create csm-allow-health-checks --allow=tcp --source-ranges=130.211.0.0/22,35.191.0.0/16

gcloud services enable \
    container.googleapis.com \
    meshconfig.googleapis.com \
    trafficdirector.googleapis.com

```

## Setup the IAM and accounts

```bash

gcloud iam service-accounts create asm --display-name asm-control

gcloud projects add-iam-policy-binding ${PROJECT_ID} --role roles/meshconfig.writer \
  --member serviceAccount:asm-control@${PROJECT_ID}.iam.gserviceaccount.com

gcloud projects add-iam-policy-binding ${PROJECT_ID} --role roles/compute.admin \
  --member serviceAccount:asm-control@${PROJECT_ID}.iam.gserviceaccount.com 

gcloud  iam service-accounts keys create google-cloud-key.json --iam-account=asm-control@costin-istio.iam.gserviceaccount.com 

```

For GKE metadata server (after IAM is fixed):

```bash

gcloud iam service-accounts add-iam-policy-binding --role roles/iam.workloadIdentityUser \
   --member "serviceAccount:${PROJECT_ID}.svc.id.goog[kube-system/glbc]" \
    asm@${PROJECT_ID}.iam.gserviceaccount.com

kubectl annotate serviceaccount --namespace kube-system glbc \
    iam.gke.io/gcp-service-account=asm@${PROJECT_ID}.iam.gserviceaccount.com
```


# GKE cluster setup

## Upgrade: Nodepool workaround

1. Create a new node pool, 

```
gcloud beta container --project $PROJECT node-pools create tdpool \
  --cluster $CLUSTER  --zone $ZONE \
  --machine-type "n1-standard-4" \
   --image-type "UBUNTU" --disk-type "pd-standard" --disk-size "100" \
  --enable-autoscaling --min-nodes "1" --max-nodes "8" \
  --scopes "https://www.googleapis.com/auth/devstorage.read_only","https://www.googleapis.com/auth/logging.write","https://www.googleapis.com/auth/monitoring","https://www.googleapis.com/auth/cloud-platform","https://www.googleapis.com/auth/servicecontrol","https://www.googleapis.com/auth/service.management.readonly","https://www.googleapis.com/auth/trace.append"  \
  --workload-metadata-from-node=EXPOSED 



# Replace GKE_METADATA_SERVER

gcloud beta container node-pools update default-pool \
  --cluster=$CLUSTER --zone ${ZONE} \
  --workload-metadata-from-node=EXPOSED
  
```

2. Delete the old pool, so workloads migrate to the new one

## New cluster


```

gcloud beta container --project $PROJECT clusters create $CLUSTER \
 --zone $ZONE --no-enable-basic-auth --release-channel "regular" \
 --machine-type "n1-standard-1" --image-type "UBUNTU" --disk-type "pd-standard" --disk-size "100" \
 --metadata disable-legacy-endpoints=true \
 --scopes "https://www.googleapis.com/auth/devstorage.read_only","https://www.googleapis.com/auth/logging.write","https://www.googleapis.com/auth/monitoring","https://www.googleapis.com/auth/cloud-platform","https://www.googleapis.com/auth/servicecontrol","https://www.googleapis.com/auth/service.management.readonly","https://www.googleapis.com/auth/trace.append"  \
--network "projects/costin-istio/global/networks/default" --subnetwork "projects/costin-istio/regions/us-central1/subnetworks/default" \
 --default-max-pods-per-node "110" \
 --enable-autoscaling --min-nodes "1" --max-nodes "8" \
 --enable-network-policy \
 --addons HorizontalPodAutoscaling,KubernetesDashboard \
  --enable-autoupgrade --enable-autorepair  \
  --identity-namespace "costin-istio.svc.id.goog"
  --num-nodes "3" --enable-stackdriver-kubernetes --enable-ip-alias \
  --workload-metadata-from-node=EXPOSED


```


# Preparing the config

Run generate.sh script - will create gce.conf and asm.conf


# Installing

1. Upload the config:

```bash

# files created by generate.sh

kubectl -n kube-system create secret generic google-cloud-key  --from-file key.json=google-cloud-key.json
kubectl -n istio-system create secret generic google-cloud-key  --from-file key.json=google-cloud-key.json
kubectl -n istio-system apply -f configmap-galley.yaml

kubectl apply -k neg 
kubectl apply -k .




```


# Checking the installation

```shell script 

kubectl -n kube-system logs -l name=gcp-lb-controller
kubectl -n kube-system delete pod -l name=gcp-lb-controller

kubectl -n istio-system logs -l app=istiod-asm
kis delete po -l app=istiod-remote
```
