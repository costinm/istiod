# Configuring Istiod webhooks using CertManager

This is a sample of using CertManager to issue certificates for
a gateway. 

It will configure both ACME issuers and a self-signed CA authority.

It needs to be installed in the same namespace with the gateway
( ex. istio-gate ) or in istio-system.

This chart will install namespace issuers:

- an ACME http challenge certificate issuer named 'acme'
- (optional) a certificate issuer named 'acme-dns' (currently only for GCP)
- 'istio-issuer' - backed by 'istio-certmanager-ca' secret

Will also create a Certificate named 'istiod' and associated secret, 
signed by 'istio-issuer'.

Finally, will install a mutating webhook for istiod, using both 
'istiod.istio-system.svc' and a custom domain name, and inject 
it with the istio-certmanager-ca root.

## Cluster wide settings.

WIP.
