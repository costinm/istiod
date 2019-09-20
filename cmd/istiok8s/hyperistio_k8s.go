package main

import (
	"flag"
	"github.com/costinm/istio-vm/pkg/istiostart"
	"github.com/costinm/istio-vm/pkg/k8s"
	"log"
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

	// Init certificates
	initCerts()

	s, err := istiostart.Init()
	if err != nil {
		log.Fatal("Failed to start ", err)
	}

	kc, err := k8s.InitK8S(s, s.Args)
	if err != nil {
		log.Fatal("Failed to start k8s", err)
	}

	kc.WaitForCacheSync(stop)

	s.Start(stop)

	s.WaitDrain(".")
}

func initCerts() {

}
