package main

import (
	"flag"
	"istio.io/istio/pkg/istiod"
	"log"
)

// hyperistio runs istio control plane components in one binary, using a directory based config by
// default or MCP sources. It is intended for testing/debugging/prototyping, as well as for running on VMs.
//
// Directory structure has a base directory, which can be mounted in a docker container or as a config map:
//
// Binaries: /usr/local/bin
// Base dir is the current working dir.
//
// conf/... - config files.
// run/... - created envoy config, run files
// certs/ - certificate directory. If found, an envoy sidecar is started for control plane using the certs
// conf/ca - root CA directory.
//
// This will start an envoy sidecar, using SDS for certificates. There is no restart capability, just drain (currently
// off, debugging /dev/shm issues)
//
//
func main() {
	stop := make(chan struct{})

	// In k8s, the config is mounted under /etc/istio/config/mesh
	// For VM, we'll use ./conf/hyperistio/mesh.yaml
	s, err := istiod.InitConfig("/etc/istio/config")
	if err != nil {
		log.Fatal("Failed to start ", err)
	}

	err = s.InitDiscovery()

	s.Start(stop, nil)

	s.WaitDrain(".")
}
