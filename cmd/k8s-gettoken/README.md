Tool to get a K8S JWT with a given audience.

Depeding on the kube config, can use:
- a 'cluster admin' credentials
- a regular KSA, with RBAC permission to the given KSA
- a GCP GSA or platform specific identity, with RBAC permission to the KSA

TODO:
- "-d" - run as a server, periodic refresh
- create an env mirroring istio-agent expected paths ( XDS CA, etc)
- labels, etc from WorkloadGroup
