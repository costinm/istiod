package main

import (
	"encoding/json"

	"istio.io/istio/pilot/pkg/bootstrap"
	"istio.io/istio/pkg/cmd"
	"istio.io/pkg/ctrlz"
	"istio.io/pkg/log"
)

// Extracted from a default server, based on the normal flags.
var DefaultFlags = `
 {
   "ServerOptions": {
      "HTTPAddr": ":8080",
      "HTTPSAddr": ":15017",
      "GRPCAddr": ":15010",
      "MonitoringAddr": ":15014",
      "EnableProfiling": true,
      "TLSOptions": {
         "CaCertFile": "",
         "CertFile": "",
         "KeyFile": "",
         "TLSCipherSuites": null,
         "CipherSuits": null
      },
      "SecureGRPCAddr": ":15012"
   },
   "InjectionOptions": {
      "InjectionDirectory": "./var/lib/istio/inject"
   },
   "PodName": "",
   "Namespace": "istio-system",
   "Revision": "grpc",
   "MeshConfigFile": "./etc/istio/config/mesh",
   "NetworksConfigFile": "/etc/istio/config/meshNetworks",
   "RegistryOptions": {
      "FileDir": "",
      "Registries": [
         "Kubernetes"
      ],
      "KubeOptions": {
         "SystemNamespace": "",
         "WatchedNamespaces": "",
         "ResyncPeriod": 60000000000,
         "DomainSuffix": "cluster.local",
         "ClusterID": "Kubernetes",
         "Metrics": null,
         "XDSUpdater": null,
         "NetworksWatcher": null,
         "MeshWatcher": null,
         "EndpointMode": 0,
         "KubernetesAPIQPS": 80,
         "KubernetesAPIBurst": 160,
         "SyncInterval": 0,
         "SyncTimeout": null,
         "DiscoveryNamespacesFilter": null
      },
      "ClusterRegistriesNamespace": "istio-system",
      "KubeConfig": "",
      "DistributionCacheRetention": 60000000000,
      "DistributionTrackingEnabled": true
   },
   "CtrlZOptions": {
      "Port": 9876,
      "Address": "localhost"
   },
   "Plugins": [
      "ext_authz",
      "authn",
      "authz"
   ],
   "KeepaliveOptions": {
      "Time": 30000000000,
      "Timeout": 10000000000,
      "MaxServerConnectionAge": 9223372036854775807,
      "MaxServerConnectionAgeGrace": 10000000000
   },
   "ShutdownDuration": 10000000000,
   "JwtRule": ""
}
`

// Istiod is an alternative startup file for pilot.
// It uses no cobra flags and doesn't have multiple functions.
//
// Defaults and configuration is done in code, env variables override.
func main() {
	serverArgs := bootstrap.NewPilotArgs(func(p *bootstrap.PilotArgs) {
		// Set Defaults
		p.CtrlZOptions = ctrlz.DefaultOptions()
		// TODO replace with mesh config?
		p.InjectionOptions = bootstrap.InjectionOptions{
			InjectionDirectory: "./var/lib/istio/inject",
		}
	})

	err := json.Unmarshal([]byte(DefaultFlags), serverArgs)

	// Explicitly set options - in Istio they are based on default values from CLI
	//serverArgs.Namespace = os.Getenv("POD_NAMESPACE")
	//serverArgs.RegistryOptions.Registries = []string{"Kubernetes"}


	// Create the server for the discovery service.
	discoveryServer, err := bootstrap.NewServer(serverArgs)
	if err != nil {
		log.Fatalf("failed to create discovery service: %v", err)
	}
	stop := make(chan struct{})

	// Start the server
	if err := discoveryServer.Start(stop); err != nil {
		log.Fatalf("failed to start discovery service: %v", err)
	}

	cmd.WaitSignal(stop)
	// Wait until we shut down. In theory this could block forever; in practice we will get
	// forcibly shut down after 30s in Kubernetes.
	discoveryServer.WaitUntilCompletion()

}
