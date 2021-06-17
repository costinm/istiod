# Configuring Istiod webhooks using CertManager

This is a sample of using CertManager to issue certificates for
gateways. The chart should be installed in istio-system ( or where istiod is running ).

It will configure both ACME issuers and a self-signed CA authority.

# DNS certs

### ClusterIssuer

- simplest
- any namespace can create a certificate request and get a certificate - so it 
MUST be secured with RBAC, only 'certificate admin' should be allowed.
- even with RBAC, you can't protect individual domains ( foo.com -> foo admin ).

Best if there is a single 'all domains owner' ( including CI/CD )

### Issuer 

It needs to be installed in the same namespace with the gateway
( ex. istio-gate ) or in istio-system, depending on secret access. 
For Istio SDS, it is istio-system.


# Workload and istiod self-signed certificates

The chart will create an Issuer "istio-issuer", in istio-system, using 
"istio-certmanager-ca" secret.

Will also create a DNS Certificate named 'istiod' and associated secret,
signed by 'istio-issuer'.

- 'istio-issuer' - backed by 'istio-certmanager-ca' secret

Finally, will install a mutating webhook for istiod, using both
'istiod.istio-system.svc' and a custom domain name, and inject
it with the istio-certmanager-ca root.
