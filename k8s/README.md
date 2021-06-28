Small library to init a Istio and Kubeconfig environment for a GCP VM, Cloudrun, or any other
environment with sufficient credentials.

Credentials can be provided by a local metadata server or downloaded service account.
The SA must have the IAM permissions to connect to the GKE cluster. This also works for 
GKE Connect Gateway, for private or non-GKE clusters registered in the hub.

