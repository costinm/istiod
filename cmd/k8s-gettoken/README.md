Tool to get a K8S JWT with a given audience.

Depeding on the kube config, can use:
- a 'cluster admin' credentials
- a regular KSA, with RBAC permission to the given KSA
- a GCP GSA or platform specific identity, with RBAC permission to the KSA
