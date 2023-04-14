# Istio agent 

This is an optional component providing the functionality of the sidecar istio agent:

- DNS server backed by istiod info
- XDS proxy - will accept requests from local node, using a K8S JWT and plain text. 
- SDS exchange - requires a K8S JWT

# Limitations

- the DNS server provides a 'view' from istio-agent namespace, no split horizon, producer oriented.
- TODO
