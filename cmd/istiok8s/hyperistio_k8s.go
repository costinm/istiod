package main

import (
	"flag"
	"github.com/costinm/istio-vm/pkg/istiostart"
	"github.com/costinm/istio-vm/pkg/k8s"
	"istio.io/istio/security/pkg/nodeagent/cache"
	"istio.io/istio/security/pkg/nodeagent/sds"
	"istio.io/istio/security/pkg/nodeagent/secretfetcher"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"os"
	"time"

	meshv1 "istio.io/api/mesh/v1alpha1"
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

	s, err := istiostart.InitConfig(13000, "./etc/istio/config")
	if err != nil {
		log.Fatal("Failed to start ", err)
	}

	// InitConfig certificates
	go initCerts(s, client, kcfg)

	kc, err := k8s.InitK8S(s, client, kcfg, s.Args)
	if err != nil {
		log.Fatal("Failed to start k8s", err)
	}

	// Initialize Galley config source for K8S.
	galleyK8S, err := kc.NewGalleyK8SSource(s.Galley.Resources)
	s.Galley.Sources = append(s.Galley.Sources, galleyK8S)

	err = s.InitDiscovery()
	if err != nil {
		log.Fatal("Failed to start ", err)
	}

	kc.InitK8SDiscovery(s, client, kcfg, s.Args)

	if false {
		kc.WaitForCacheSync(stop)
	}

	StartSDSK8S(".", s.Mesh)

	err = s.Start(stop, kc.OnXDSStart)
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

// Start the workload SDS server. Will run on the UDS path - Envoy sidecar will use a cluster
// to expose the UDS path over TLS, using Apiserver-signed certs.
// SDS depends on k8s.
func StartSDSK8S(baseDir string, config *meshv1.MeshConfig) error {

	// This won't work on VM - only on K8S.
	var sdsCacheOptions cache.Options
	var serverOptions sds.Options

	// Compat with Istio env
	caProvider := os.Getenv("CA_PROVIDER")
	if caProvider == "" {
		caProvider = "Citadel"
	}

	wSecretFetcher, err := secretfetcher.NewSecretFetcher(false,
		serverOptions.CAEndpoint, caProvider, true,
		[]byte(""), "", "", "", "")
	if err != nil {
		log.Fatal("failed to create secretFetcher for workload proxy", err)
	}
	sdsCacheOptions.TrustDomain = serverOptions.TrustDomain
	sdsCacheOptions.RotationInterval = 5 * time.Minute
	serverOptions.RecycleInterval = 5 * time.Minute
	serverOptions.EnableWorkloadSDS = true

	sdsCacheOptions.Plugins = sds.NewPlugins(serverOptions.PluginNames)
	workloadSecretCache := cache.NewSecretCache(wSecretFetcher, sds.NotifyProxy, sdsCacheOptions)

	// GatewaySecretCache loads secrets from K8S
	_, err = sds.NewServer(serverOptions, workloadSecretCache, nil)

	if err != nil {
		log.Fatal("Failed to start SDS server", err)
	}

	return nil
}
