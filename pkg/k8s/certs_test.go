package k8s

import (
	kubelib "istio.io/istio/pkg/kube"
	"testing"
)

func TestCerts(t *testing.T) {
	client, err := kubelib.CreateClientset("", "")
	if err != nil {
		t.Fatal("Missing K8S", err)
	}

	certChain, keyPEM, err := GenKeyCertK8sCA(client.CertificatesV1beta1(), "istio-pilot.istio-system")
	if err != nil {
		t.Fatal("Fail to generate cert", err)
	}

	t.Log("Key", string(keyPEM))

	// Include the root cert in the chain
	caCert, err := readCACert("")
	if err != nil {
		t.Fatal("Fail to generate cert", err)
	}
	certChain = append(certChain, caCert...)

	t.Log("Cert Chain: ", string(certChain))
	t.Log("CA:", string(caCert))

}
