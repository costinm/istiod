Signs SSH certificates for host and client.

This is similar with Istiod/Citadel signing of workload certificates,
but for SSH certificates.

Protocol is a bit simplified, using a POST request instead of a proto.
In practice Istio requests take a PEM CSR as param and return a list of certs.
For SSH, the input is the public key and response is the certificate.

Authentication uses same K8S JWT token as Istiod-CA.

Host certs use the canonical service, namespace and domain.
User certs use the KSA, namespace and domain.

# Configuration

The app is intended to run in KNative, CloudRun or similar env.

The setup uses env variables:

Configuring the K8S server - required when running in CloudRun
- CLUSTER
- LOCATION
- PROJECT

