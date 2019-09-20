package main

import (
	"flag"
	"github.com/costinm/istio-vm/pkg/istiostart"
	"github.com/costinm/istio-vm/pkg/k8s"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"os"
)

// Istio control plane with K8S support.
//
// - config is loaded from local K8S (in-cluster or using KUBECONFIG)
// - local endpoints and additional registries from K8S
// - additional MCP registries supported
// - includes a Secret controller that provisions certs as secrets.
//
// Normal hyperistio is using local config files and MCP sources for config/endpoints,
// as well as SDS backed by a file-based CA.
func main() {
	flag.Parse()
	stop := make(chan struct{})

	client, kcfg, err := k8s.CreateClientset(os.Getenv("KUBECONFIG"), "")
	if err != nil {
		log.Fatal("Failed to connect to k8s", err)
	}

	s, err := istiostart.Init(13000)
	if err != nil {
		log.Fatal("Failed to start ", err)
	}

	// Init certificates
	go initCerts(s, client, kcfg)

	kc, err := k8s.InitK8S(s, client, kcfg, s.Args)
	if err != nil {
		log.Fatal("Failed to start k8s", err)
	}

	if false {
		kc.WaitForCacheSync(stop)
	}

	err = s.Start(stop)
	if err != nil {
		log.Fatal("Failure on start", err)
	}

	s.WaitDrain(".")
}

func initCerts(server *istiostart.Server, client *kubernetes.Clientset, cfg *rest.Config) {
	// TODO: fallback to citadel (or custom CA)

	certChain, keyPEM, err := k8s.GenKeyCertK8sCA(client.CertificatesV1beta1(), "istio-system",
		"istio-pilot.istio-system")
	if err != nil {
		log.Fatal("Failed to initialize certs")
	}
	server.CertChain = certChain
	server.CertKey = keyPEM

}
