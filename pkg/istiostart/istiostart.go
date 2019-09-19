package istiostart

import (
	"istio.io/istio/pilot/pkg/features"
	"log"
	"os"
	"time"

	meshv1 "istio.io/api/mesh/v1alpha1"
	agent "istio.io/istio/pkg/bootstrap"
)

// Should be called at the end, if we receive SIGINT or SIGTERM
func DrainEnvoy(base string, cfg *meshv1.ProxyConfig) {
	// Simplified version:
	// - hot restart envoy with new config
	// - sleep terminationDrainDuration
	// - exit

	stop := make(chan error)
	//features.EnvoyBaseId.DefaultValue = "1"
	process, err := agent.RunProxy(cfg, "nodeid", 2,
		base+"/conf/sidecar/envoy_bootstrap_drain.json", stop,
		os.Stdout, os.Stderr, []string{
			// "--disable-hot-restart",
			// "-l", "trace",
		})

	if err != nil {
		log.Fatal("Failed to drain, abrupt termination", err)
	}

	go func() {
		// Should not happen.
		process.Wait()
		log.Fatal("Envoy terminated after drain")
	}()

	// Env variable from Istio
	time.Sleep(features.TerminationDrainDuration())
}
