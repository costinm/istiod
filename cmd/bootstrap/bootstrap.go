package main

import (
	"fmt"
	"github.com/gogo/protobuf/types"
	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pkg/bootstrap"
	"os"
	"time"
)

// Generate an envoy config file, using Istio bootstrap package
// This can be used with upstream envoy, without pilot-agent.

// TODO: this should move inside hyperistio, behind a URL - generation should just curl with the env variables as params

const (
	sdsPath = "/var/run/sds/uds_path"
)

func main() {
	xds := os.Getenv("XDS_ADDR")
	if xds == "" {
		xds = "https://istio-pilot.istio-system:15011"
	}

	// We need a proxy config - either from pilot or as a file.
	// If the proxy config is present, use it. Otherwise attempt to find the pilot address.
	//
	// To connect, we should assume a k8s-signed DNS cert and k8s pub key in the well-known location. This can be
	// emulated in docker/VM -

	instanceIp := os.Getenv("INSTANCE_IP")

	// New
	trustDomain := os.Getenv("TRUST_DOMAIN")
	if trustDomain == "" {
		trustDomain = "cluster.local"
	}

	// ISTIO_BOOTSTRAP env can override the CustomConfigFile
	//
	// Inside the template:
	// config - the ProxyConfig and derived config (connect_timeout, etc)
	// pilot_SAN - the san array
	// nodeID - second param
	// meta - from ENV, including JSON type
	//

	bootstrap.WriteBootstrap(&meshconfig.ProxyConfig{
		DiscoveryAddress: "localhost:15010",
		ConfigPath:       "conf/sidecar/gen_bootstrap",
		BinaryPath:       "/usr/local/bin/envoy",
		ServiceCluster:   "test",
		CustomConfigFile: "conf/sidecar/envoy_bootstrap_v2.json", // input template
		ConnectTimeout:   types.DurationProto(5 * time.Second),   // crash if not set
		DrainDuration:    types.DurationProto(30 * time.Second),  // crash if 0
		StatNameLength:   189,
	},
		fmt.Sprintf("sidecar~%s~a~a", instanceIp),
		0,
		[]string{
			"istio-pilot.istio-system",
			fmt.Sprintf("spiffe://%s/ns/%s/sa/%s", trustDomain, "istio-system", "istio-pilot-service-account"),
		},
		map[string]interface{}{},
		[]string{},
		[]string{},
		"60s")

}
