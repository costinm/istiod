package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gogo/protobuf/types"

	meshconfig "istio.io/api/mesh/v1alpha1"

	"istio.io/istio/galley/pkg/server"
	"istio.io/istio/galley/pkg/server/settings"
	"istio.io/istio/pilot/pkg/bootstrap"
	"istio.io/istio/pilot/pkg/proxy/envoy"
	"istio.io/istio/pilot/pkg/serviceregistry"
	agent "istio.io/istio/pkg/bootstrap"
	"istio.io/istio/pkg/config/mesh"
	"istio.io/istio/pkg/keepalive"
)

var (
	runEnvoy = flag.Bool("envoy", true, "Start envoy")
)

// hyperistio runs istio control plane components in one binary, using a directory based config by
// default. It is intended for testing/debugging/prototyping, as well as for running on VMs.
//
// Directory structure:
//
// Binaries: /usr/local/bin
// Base dir is the current working dir.
//
// conf/... - config files.
// run/... - created envoy config, run files
//
// certs/ - certificate directory. If found, an envoy sidecar is started for control plane using the certs
// conf/ca - root CA directory.

func main() {
	flag.Parse()

	err := startAll()
	if err != nil {
		log.Fatal("Failed to start ", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

}

func startAll() error {
	binDir := "/usr/local/bin"
	baseDir := "./"
	err := startGalley(baseDir)
	if err != nil {
		return err
	}
	err = startPilot(binDir, baseDir)
	if err != nil {
		return err
	}

	if *runEnvoy {
		err = startEnvoy(binDir, baseDir)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: use pilot-agent code, and refactor it to extract the core functionality.
func startEnvoy(binDir, baseDir string) error {
	os.Mkdir(baseDir+"/run", 0700)
	cfg := &meshconfig.ProxyConfig{
		DiscoveryAddress: "localhost:12010",
		ConfigPath:       baseDir + "/run",
		BinaryPath:       binDir + "/envoy",
		ServiceCluster:   "istio",
		CustomConfigFile: baseDir + "/conf/sidecar/envoy_bootstrap_v2.json",
		ConnectTimeout:   types.DurationProto(5 * time.Second),  // crash if not set
		DrainDuration:    types.DurationProto(30 * time.Second), // crash if 0
		StatNameLength:   189,
	}
	nodeId := "sidecar~127.0.0.2~a~a"
	cfgF, err := agent.WriteBootstrap(cfg, nodeId, 1, []string{}, nil, os.Environ(), []string{}, "60s")
	if err != nil {
		return err
	}
	stop := make(chan error)
	envoyLog, err := os.Create(baseDir + "/run/envoy_hyperistio_sidecar.log")
	if err != nil {
		envoyLog = os.Stderr
	}
	_, err = agent.RunProxy(cfg, nodeId, 1, cfgF, stop, envoyLog, envoyLog, []string{
		"--disable-hot-restart", // "-l", "trace",
	})
	return err
}

// startPilot with defaults:
// - http port 15007
// - grpc on 15010
// - grpcs in 15011 - certs from PILOT_CERT_DIR or ./tests/testdata/certs/pilot
// - mixer set to localhost:9091 (runs in-process),
//-  http proxy on 15002 (so tests can be run without iptables)
//- config from $ISTIO_CONFIG dir (defaults to in-source tests/testdata/config)
func startPilot(binDir, baseDir string) error {
	stop := make(chan struct{})

	mcfg := mesh.DefaultMeshConfig()
	mcfg.ProxyHttpPort = 12002

	// Load config from the in-process Galley.
	// We can also configure Envoy to listen on 9901 and galley on different port, and LB
	mcfg.ConfigSources = []*meshconfig.ConfigSource{
		&meshconfig.ConfigSource{
			Address: "localhost:12901",
		},
	}

	// Create a test pilot discovery service configured to watch the tempDir.
	args := bootstrap.PilotArgs{
		Namespace: "testing",
		DiscoveryOptions: envoy.DiscoveryServiceOptions{
			HTTPAddr:        ":12007",
			GrpcAddr:        ":12010",
			SecureGrpcAddr:  ":12011",
			EnableCaching:   true,
			EnableProfiling: true,
		},

		Mesh: bootstrap.MeshArgs{

			MixerAddress:    "localhost:9091",
			RdsRefreshDelay: types.DurationProto(10 * time.Millisecond),
		},
		Config: bootstrap.ConfigArgs{},
		Service: bootstrap.ServiceArgs{
			// Using the Mock service registry, which provides the hello and world services.
			Registries: []string{
				string(serviceregistry.MCPRegistry)},
		},

		// MCP is messing up with the grpc settings...
		MCPMaxMessageSize:        1024 * 1024 * 64,
		MCPInitialWindowSize:     1024 * 1024 * 64,
		MCPInitialConnWindowSize: 1024 * 1024 * 64,

		MeshConfig:       &mcfg,
		KeepaliveOptions: keepalive.DefaultOption(),
	}

	bootstrap.FilepathWalkInterval = 5 * time.Second

	log.Println("Using mock configs: ")

	// Create and setup the controller.
	s, err := bootstrap.NewServer(args)
	if err != nil {
		return err
	}

	// Start the server.
	if err := s.Start(stop); err != nil {
		return err
	}
	return nil
}

// Start the galley component, with its args.

func startGalley(baseDir string) error {
	args := settings.DefaultArgs()
	args.ConfigPath = baseDir + "/conf/istio/simple"
	// TODO: load a json file to override defaults (for all components)

	args.ValidationArgs.EnableValidation = false
	args.ValidationArgs.EnableReconcileWebhookConfiguration = false
	args.APIAddress = "tcp://0.0.0.0:12901"
	args.Insecure = true
	args.EnableServer = true
	args.DisableResourceReadyCheck = true
	args.MeshConfigFile = baseDir + "/conf/pilot/mesh.yaml"
	args.MonitoringPort = 12015

	gs := server.New(args)
	err := gs.Start()
	if err != nil {
		log.Fatalln("Galley startup error", err)
	}

	return nil
}
