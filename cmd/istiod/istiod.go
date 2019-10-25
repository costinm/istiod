package main

import (
	"io/ioutil"
	"istio.io/istio/pkg/istiod"
	"istio.io/istio/pkg/istiod/k8s"
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

// Istiod variant - with a subset of the features.
// The istiod in istio.io has the full feature set, this attempts to make sure subsets can still be produced.
func main() {
	stop := make(chan struct{})

	// First create the k8s clientset - and return the config source.
	// The config includes the address of apiserver and the public key - which will be used
	// after cert generation, to check that Apiserver-generated certs have same key.
	client, kcfg, err := k8s.CreateClientset(os.Getenv("KUBECONFIG"), "")
	if err != nil {
		// TODO: 'local' mode where k8s is not used - using the config.
		log.Fatal("Failed to connect to k8s", err)
	}

	// Load the mesh config. Note that the path is slightly changed - attempting to move all istio
	// related under /var/lib/istio, which is also the home dir of the istio user.
	istiods, err := istiod.NewIstiod(kcfg, client, "/var/lib/istio/config")
	if err != nil {
		log.Fatal("Failed to start istiod ", err)
	}

	// Create k8s-signed certificates. This allows injector, validation to work without Citadel, and
	// allows secure SDS connections to Istiod.
	initCerts(istiods, client, kcfg)

	// Init k8s related components: Pilot discovery, MC discovery. Code kept in separate package.
	// Temp: also Pilot config watcher, until we optimize galley integration
	k8sServer, err := k8s.InitK8S(istiods, client, kcfg, istiods.Args)
	if err != nil {
		log.Fatal("Failed to start k8s controllers ", err)
	}

	err = istiods.InitDiscovery()
	if err != nil {
		log.Fatal("Failed to init XDS server ", err)
	}

	k8sServer.InitK8SDiscovery(istiods, kcfg, istiods.Args)

	err = istiods.Start(stop, k8sServer.OnXDSStart)
	if err != nil {
		log.Fatal("Failure on start XDS server", err)
	}

	// Injector should run along, even if not used - but only if the injection template is mounted.
	if _, err := os.Stat("./var/lib/istio/inject/injection-template.yaml"); err == nil {
		err = k8s.StartInjector(stop)
		if err != nil {
			log.Fatalf("Failure to start injector ", err)
		}
	}

	// Options based on the current 'defaults' in istio.
	// If adjustments are needed - env or mesh.config ( if of general interest ).

	istiod.RunCA(istiods.GrpcServer, client, &istiod.CAOptions{
		TrustDomain: istiods.Mesh.TrustDomain,
		DualUse: true,
		SelfSignedCA: true, // for existing CA - mount a secret in the expected location
		IstioNamespace: "istio-system",
		MaxWorkloadCertTTL: 90 * 24 * time.Hour,
		WorkloadCertTTL: 90 * 24 * time.Hour,
	})

	istiods.Serve(stop)
	istiods.WaitStop(stop)
}

func initCerts(server *istiod.Server, client *kubernetes.Clientset, cfg *rest.Config) {
	// TODO: fallback to citadel (or custom CA)

	certChain, keyPEM, err := k8s.GenKeyCertK8sCA(client.CertificatesV1beta1(), "istio-system",
		"istio-pilot.istio-system,istiod.istio-system")
	if err != nil {
		log.Fatal("Failed to initialize certs")
	}
	server.CertChain = certChain
	server.CertKey = keyPEM

	// Save the certificates to /var/run/secrets/istio-dns
	os.MkdirAll(istiod.DNSCertDir, 0700)
	err = ioutil.WriteFile(istiod.DNSCertDir+"/key.pem", keyPEM, 0700)
	if err != nil {
		log.Fatal("Failed to write certs", err)
	}
	err = ioutil.WriteFile(istiod.DNSCertDir+"/cert-chain.pem", certChain, 0700)
	if err != nil {
		log.Fatal("Failed to write certs")
	}
}

// Start the workload SDS server. Will run on the UDS path - Envoy sidecar will use a cluster
// to expose the UDS path over TLS, using Apiserver-signed certs.
// SDS depends on k8s.
//
// TODO: modify NewSecretFetcher, add method taking a kube client ( for consistency )
// TODO: modify NewServer, add method taking a grpcServer
func StartSDSK8S(baseDir string, config *meshv1.MeshConfig) error {

	// This won't work on VM - only on K8S.
	var sdsCacheOptions cache.Options
	var serverOptions sds.Options

	// Compat with Istio env - will determine the plugin used for connecting to the CA.
	caProvider := os.Getenv("CA_PROVIDER")
	if caProvider == "" {
		caProvider = "Citadel"
	}

	// Compat with Istio env
	// Will use istio-system/istio-security config map to load the root CA of citadel.
	// The address can be the Gateway address for the cluster
	// TODO: load the GW address of the cluster to auto-configure
	// TODO: Citadel should also use k8s-api signed certificates ( possibly on different port )
	caAddr := os.Getenv("CA_ADDR")
	if caAddr == "" {
		// caAddr = "istio-citadel.istio-system:8060"
		// For testing with port fwd (kfwd istio-system istio=citadel 8060:8060)
		caAddr = "localhost:8060"
	}

	wSecretFetcher, err := secretfetcher.NewSecretFetcher(false,
		caAddr, caProvider, true,
		[]byte(""), "", "", "", "")
	if err != nil {
		log.Fatal("failed to create secretFetcher for workload proxy", err)
	}
	sdsCacheOptions.TrustDomain = serverOptions.TrustDomain
	sdsCacheOptions.RotationInterval = 5 * time.Minute
	serverOptions.RecycleInterval = 5 * time.Minute
	serverOptions.EnableWorkloadSDS = true
	serverOptions.WorkloadUDSPath = "./sdsUDS"
	sdsCacheOptions.Plugins = sds.NewPlugins(serverOptions.PluginNames)
	workloadSecretCache := cache.NewSecretCache(wSecretFetcher, sds.NotifyProxy, sdsCacheOptions)

	// GatewaySecretCache loads secrets from K8S - they should be in same namespace with ingress gateway, will use
	// standalone node agent.
	_, err = sds.NewServer(serverOptions, workloadSecretCache, nil)

	if err != nil {
		log.Fatal("Failed to start SDS server", err)
	}

	return nil
}
