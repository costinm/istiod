package bootstrap

import (
	"bytes"
	"io/ioutil"
	"net"
	"os"
	"text/template"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	xdscreds "google.golang.org/grpc/credentials/xds"
	"google.golang.org/grpc/xds"
)

// Dubo: https://github.com/apache/dubbo-samples/blob/master/3-extensions/registry/dubbo-samples-xds/dubbo-samples-xds-provider/src/main/resources/spring/dubbo-provider.properties
// dubbo.registry.address=xds://istiod.istio-system.svc:15012
// Also supports k8s, consul, etc
// Impl: https://github.com/apache/dubbo/tree/3.3/dubbo-xds/src/main/java/org/apache/dubbo/registry/xds/istio

const grpcTemplate = `
{
  "xds_servers": [
    {
      "server_uri": "{{ .opts.DiscoveryAddress}}",
      "channel_creds": [{"type": "insecure"}],
      "server_features" : ["xds_v3"]
    }
  ],
  "node": {
    "id": "sidecar~{{.opts.IP}}~{{.opts.Name}}.{{.opts.Namespace}}~{{.opts.Namespace}}.cluster.local",
    "metadata": {
      "INSTANCE_IPS": "{{.opts.InstanceIPS}}",
      "PILOT_SAN": [
        "istiod.istio-system.svc"
      ],
      "GENERATOR": "grpc",
      "NAMESPACE": "{{.opts.Namespace}}"
    },
    "localisty": {},
    "UserAgentVersionType": "istiov1"
  },
  {{if .opts.CertDir}}
  "certificate_providers": {
    "default": {
      "plugin_name": "file_watcher",
      "config": {
        "certificate_file": "{{.opts.CertDir}}/cert-chain.pem",
        "private_key_file": "{{.opts.CertDir}}/key.pem",
        "ca_certificate_file": "{{.opts.CertDir}}/root-cert.pem",
        "refresh_interval": "900s"
      }
    }
  },
  {{end}}
  "server_listener_resource_name_template": "xds.istio.io/grpc/lds/inbound/%s"
}
`

type GenerateBootstrapOptions struct {
	Name string

	NodeMetadata map[string]interface{}

	DiscoveryAddress string

	CertDir string

	Namespace string

	// 'primary' IP address
	IP string

	// Comma separated list of all IPs
	InstanceIPS string

	GRPCOptions []grpc.ServerOption
}

// GRPCServer is the interface implemented by both grpc
type GRPCServer interface {
	RegisterService(*grpc.ServiceDesc, interface{})
	Serve(net.Listener) error
	Stop()
	GracefulStop()
	GetServiceInfo() map[string]grpc.ServiceInfo
}

// Istio-agent present - use it.
const defaultXDSProxy = "/etc/istio/proxy/XDS"
const certsDir = "/var/run/secrets/workload-spiffe-credentials"

// GenerateBootstrap will write a Istio bootstrap file in the location expected by gRPC, using
// Istio environment variables:
//
// XDS_ADDR - the address of the XDS server, defaults to istiod.istio-system.svc:15010 if cert not set, and 15012 if root cert found
// POD_NAMESPACE, LABELS - based on standard mounts
// ISTIO_META_env variables used like in regular Istio
// ...
func GenerateGRPCXDS(opts *GenerateBootstrapOptions) (GRPCServer, error) {
	bootF := os.Getenv("GRPC_XDS_BOOTSTRAP")
	if bootF == "" {
		return grpc.NewServer(opts.GRPCOptions...), nil
	}

	if opts == nil {
		opts = &GenerateBootstrapOptions{}
	}
	if opts.DiscoveryAddress == "" {
		if _, err := os.Stat(defaultXDSProxy); !os.IsNotExist(err) {
			opts.DiscoveryAddress = defaultXDSProxy
		}
	}
	if opts.DiscoveryAddress == "" {
		xdsAddr := os.Getenv("XDS_ADDR")
		if xdsAddr == "" {
			xdsAddr = "istiod.istio-system.svc:15010"
		}
		opts.DiscoveryAddress = xdsAddr
	}

	// TODO: if it doesn't exist, don't initialize it
	if opts.CertDir == "" {
		opts.CertDir = certsDir
	}

	creds, _ := xdscreds.NewServerCredentials(xdscreds.ServerOptions{
		FallbackCreds: insecure.NewCredentials()})
	opts.GRPCOptions = append(opts.GRPCOptions, grpc.Creds(creds))

	if opts.IP == "" {
		opts.IP = "127.0.0.3"
		opts.InstanceIPS = "127.0.0.3"
	}

	if opts.Namespace == "" {
		opts.Namespace = "default"
	}

	if opts.Name == "" {
		opts.Name = "default"
	}

	t := template.New("grpc")
	_, err := t.Parse(grpcTemplate)
	if err != nil {
		return nil, err
	}
	out := &bytes.Buffer{}

	t.Execute(out, map[string]interface{}{"opts": opts})

	if _, err := os.Stat(bootF); os.IsNotExist(err) {
		// TODO: write the bootstrap file.
		err := ioutil.WriteFile(bootF, out.Bytes(), 0700)
		if err != nil {
			return nil, err
		}
	}

	return xds.NewGRPCServer(opts.GRPCOptions...), nil
}
