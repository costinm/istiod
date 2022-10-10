gcloud alpha network-services meshes import <MESH_NAME>    --source=mesh.yaml     --location=global

gcloud run deploy <SERVICE_NAME> --image=gcr.io/<YOUR_PROJECT_ID>/<APPLICATION_TAG_NAME>:latest \
   --port=<PORT> --execution-environment=gen2 --service-mesh=td --td-mesh-name=<mesh-name>  \
   --mesh-project=<mesh-project> --service-account=<SERVICE_ACCOUNT> \
   [--custom-audiences '["<HOSTNAME>"]'] -–project=<YOUR_PROJECT_ID>


# Multiple routes possible
gcloud compute network-endpoint-groups create <NEG_NAME> \
  --region=us-central1 \
  --network-endpoint-type=serverless \
  --cloud-run-service=<SERVICE_NAME>

gcloud compute backend-services create <BS_NAME> \
  --global \
  --load-balancing-scheme=INTERNAL_SELF_MANAGED

# Can add multiple NEGs
gcloud compute backend-services add-backend <BS_NAME> \
  --global \
  --network-endpoint-group=<NEG_NAME> \
  --network-endpoint-group-region=us-central1




cat <<EOF > route.yaml
name: <ROUTE_NAME>
hostnames:
# The short name to dial to access the cloud run service, recommend setting it to
# the cloud run service name. Must match audience
- <HOSTNAME>
meshes:
- projects/<MESH_PROJECT_NUMBER>/locations/global/meshes/<MESH_NAME>
rules:
- action:
    destinations:
    - serviceName: "projects/<ROUTE_PROJECT_NUMBER>/locations/global/backendServices/<BS_NAME>"
EOF

gcloud alpha network-services http-routes import <ROUTE_NAME> \
  --source=route.yaml \
  --location=global

