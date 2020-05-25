# Istio certificates and where to find them

## K8s

Before searching for Istio certificates - it helps to understand how K8S is doing the same thing.

- https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-certs/
    - Base directory: /etc/kubernetes/pki or --cert-dir or ClusterConfiguration.certificatesDir
    - file names: ca.crt, ca.key
    - If they exist, kubeadm will not override them 
    - 'external CA' mode uses the same files
    - Possible to run the 'signer' in a separate VM --controlers=csrsigner. In this case, ca.key is not needed 
    on apiserver, ca.crt is still distributed 
    - client certificates expiration: 1 year
    - certificate renewal during upgrade !
- https://kubernetes.io/docs/setup/best-practices/certificates/
    - default CN is 'kubernetes-ca'
    - manually generate certificates, O=system:masters

K8S CA must be activated, flags --cluster-signing-cert-file --cluster-signing-key-file or:

```yaml
apiVersion: kubeadm.k8s.io/v1beta2
  kind: ClusterConfiguration
  controllerManager:
    extraArgs:
      cluster-signing-cert-file: /etc/kubernetes/pki/ca.crt
      cluster-signing-key-file: /etc/kubernetes/pki/ca.key
```

K8S uses "CN" for identity, and "O" as a 'group'. Example:  CN=system:node:<nodeName>, O=system:nodes

TODO: For 1.6, we should increase consistency and use the same naming scheme and file names in Istio.

TODO: we should use 'istio-ca' as default name for the istiod CA


## CertManager

Another popular tool is CertManager - we use it with Istio to get ACME certificates.

In CA mode, it uses a secret named "ca-key-pair", with tls.crt and tls.key files containing the root.
The "Issuer/ClusterIssuer" CR includes the secret name - so multiple CA roots can be run in same namespace.

External issuers are supported, they need to watch CertificateRequest resource and sign it.

CertificateRequest is similar - but more powerful - than K8S equivalent.
- duration
- issuer ref - can specify any issuer
- csr, isCA, usages

Status is updated with 'certificate' and 'ca'.

Certificate resource doesn't require CSR:
- subject, commonName, organization
- duration, renewBefore
- dnsNames
- ipAddresses
- uriSANs
- usages
- isCA
- issuer 
- 'secretName' defines the output, a Secret will be created.

ACME will accept only a subset.

Secret will use pkcs1 by default, pkcs8 also allowed.

Inside the secret, tls.key and tls.crt are used.

TODO: for 1.6, we should support secrets with those names. The Ingress SDS implementation is already accepting them,
for consistency we should accept them for sidecar secrets.

## Istiod 

1. If ./etc/cacerts is found, it will be used as root CA. Supports intermediary CAs. 
Works on VM too, but the file must be present. 

istio_ca.go / ROOT_CA_DIR / ./etc/cacerts

ca-key.pem - the key
ca-cert.pem - the intermediary or root cert
cert-chain.pem - root cert

2. Else, K8S secret is generated with a self-signed key. Requires K8S client.

## Istiod-grpc

Secrets for debugging will be saved in:
./var/run/secrets/istio-dns/{key.pem,cert-chain.pem}

## Sidecar

Connection to XDS server authenticated using:

./etc/istio/citadel-ca-cert/root-cert.pem


PROV_CERTS used to load initial certificate
OUTOUT_CERTS (default "", can be set to /var/run/secrets/istio)
