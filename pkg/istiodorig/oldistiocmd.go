package istiod

import (
	"istio.io/istio/security/pkg/nodeagent/cache"
	"istio.io/istio/security/pkg/nodeagent/sds"
	"istio.io/istio/security/pkg/nodeagent/secretfetcher"
	"time"
)

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

