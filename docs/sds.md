
# Protocols

- Citadel: 

Will use K8S to load cert - NAMESPACE env variable (default istio-system) used to load a "istio-security"
config map, with a "caTLSRootCert" key.

Has TODO to redial - if connection interrupted it'll be stuck. Same for the other.
istio.v1.auth.IstioCertificateService


- GoogleCA: uses IstioCertificateRequest (csr->cert_chain),  with Bearer JWT token.
Endpoint is a https:// with DNS public cert.

google.security.istioca.v1alpha1.IstioCertificateService


- Old node agent - IstioCAService.HandleCSR(), with node_agent_credential type GCP/AWS/etc

Metadata: 
- authorization - K8S-signed token
- x-goog-request-params location=zone
- 

- 
- Vault: currently not supported, missing JWT auth with audience in Vault.
