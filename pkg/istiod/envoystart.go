package istiod

import (
	"encoding/pem"
	"fmt"
	"github.com/gogo/protobuf/types"
	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// TODO: use pilot-agent code, and refactor it to extract the core functionality.

// TODO: better implementation for 'drainFile' config - used by agent.terminate()

// startEnvoy starts the envoy sidecar for Istio control plane, for TLS and load balancing.
// Should be called after cert generation
func (s *Server) StartEnvoy(baseDir string, mcfg *meshv1.MeshConfig) error {
	os.Mkdir(baseDir+"/etc/istio/proxy", 0700)

	cfg := &meshv1.ProxyConfig{}
	// Copy defaults
	pcval, _ := mcfg.DefaultConfig.Marshal()
	cfg.Unmarshal(pcval)

	// This is the local envoy serving the control plane - gets configs from localhost, no TLS
	cfg.DiscoveryAddress = fmt.Sprintf("localhost:%d", s.basePort+10)

	cfg.ProxyAdminPort = s.basePort

	// Override shutdown, it's too slow
	cfg.ParentShutdownDuration = types.DurationProto(5 * time.Second)

	cfg.ConfigPath = baseDir + "/etc/istio/proxy"

	// Let's try to use the same bootstrap that sidecars are using - without a special config for istiod.
	// Will use localhost and get configs from pilot.
	cfg.CustomConfigFile = baseDir + "/var/lib/istio/envoy/envoy_bootstrap_tmpl.json"

	nodeId := "sidecar~127.0.0.1~istio-pilot.istio-system~istio-system.svc.cluster.local"
	env := os.Environ()
	env = append(env, "ISTIO_META_ISTIO_VERSION=1.4")

	//cfgF, err := agent.WriteBootstrap(cfg, nodeId, 1, []string{
	//	"istio-pilot.istio-system",
	//	fmt.Sprintf("spiffe://%s/ns/%s/sa/%s", trustDomain, "istio-system", "istio-pilot-service-account"),
	//},
	//	map[string]interface{}{},
	//	env,
	//	[]string{"127.0.0.1"}, // node IPs
	//	"60s")

	cfgF, err := agent.New(agent.Config{
		Node:                "sidecar~127.0.0.1~istio-pilot.istio-system~istio-system.svc.cluster.local",
		DNSRefreshRate:      "300s",
		Proxy:               cfg,
		PilotSubjectAltName: nil,
		MixerSubjectAltName: nil,
		LocalEnv:            os.Environ(),
		NodeIPs:             []string{"127.0.0.1"},
		PodName:             "istiod",
		PodNamespace:        "istio-system",
		PodIP:               net.IP([]byte{127, 0, 0, 1}),
		SDSUDSPath:          "",
		SDSTokenPath:        "",
		ControlPlaneAuth:    false,
		DisableReportCalls:  true,
	}).CreateFileForEpoch(0)
	if err != nil {
		return err
	}
	log.Println("Created ", cfgF)

	// Start Envoy, using the pre-generated config. No restarts: if it crashes, we exit.
	stop := make(chan error)
	//features.EnvoyBaseId.DefaultValue = "1"
	process, err := agent.RunProxy(cfg, nodeId, 1, cfgF, stop,
		os.Stdout, os.Stderr, []string{
			"--disable-hot-restart",
			// "-l", "trace",
		})
	if err != nil {
		log.Fatal("Failed to start envoy sidecar for istio", err)
	}
	go func() {
		// Should not happen.
		process.Wait()
		log.Fatal("Envoy terminated, restart.")
	}()
	return err
}


// ReadCACert reads and check whether it is a valid certificate.
func ReadCACert(caCertPath string) ([]byte, error) {
	if caCertPath == "" {
		caCertPath = defaultCA
	}
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Errorf("failed to read CA cert, cert. path: %v, error: %v", caCertPath, err)
		return nil, fmt.Errorf("failed to read CA cert, cert. path: %v, error: %v", caCertPath, err)
	}

	b, _ := pem.Decode(caCert)
	if b == nil {
		return nil, fmt.Errorf("could not decode pem")
	}
	if b.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("ca certificate contains wrong type: %v", b.Type)
	}
	if _, err := x509.ParseCertificate(b.Bytes); err != nil {
		return nil, fmt.Errorf("ca certificate parsing returns an error: %v", err)
	}

	return caCert, nil
}

