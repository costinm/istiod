package k8s

import (
	"testing"
)

func TestCerts(t *testing.T) {

	client, kcfg, err := CreateClientset("", "")
	if err != nil {
		t.Fatal("Missing K8S", err)
	}

	certChain, keyPEM, err := GenKeyCertK8sCA(client.CertificatesV1beta1(), "istio-pilot.istio-system")
	if err != nil {
		t.Fatal("Fail to generate cert", err)
	}

	t.Log("Key\n", string(keyPEM))

	caCert := kcfg.TLSClientConfig.CAData
	certChain = append(certChain, caCert...)

	t.Log("Cert Chain:\n", string(certChain))
}
