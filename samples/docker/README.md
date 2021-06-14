Istio apps and gateway running in a single docker container.

This was originally (long, long time ago) intended for testing iptables
and VM. It is based on Istio for VM support.

Primary target is the base docker image, containing the installed proxy and a modified
startup script. This can be used as a base, adding your application.

# CloudRun

The docker images also work in CloudRun - or similar docker-based execution environments.

The images use HBONE to expose TCP ports and support e2e mTLS. 

# Limitations/missing

- ./etc/istio/pod/annotations, labels
- ...
