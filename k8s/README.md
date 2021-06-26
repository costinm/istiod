Small command line tool to init a Kubeconfig file for GCP VM, Cloudrun, or any other
environment with a metadata server producing GCP SA tokens.

The SA must have the IAM permissions to connect to the GKE cluster. This also works for 
GKE Connect Gateway, for non-GKE clusters registered in the hub.
