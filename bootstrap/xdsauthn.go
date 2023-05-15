package bootstrap

// Experimental support for custom auth for google grpc golang.

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/tls/certprovider"
	"google.golang.org/grpc/xds/bootstrap"
)

func init() {
	bootstrap.RegisterCredentials(&XDSCreds{})
	certprovider.Register(&MeshCertProvider{})
}

type MeshCertProvider struct {
}

type MeshCerts struct {
}

func (c *MeshCerts) Close() {
}

func (c *MeshCerts) KeyMaterial(ctx context.Context) (*certprovider.KeyMaterial, error) {
	//TODO implement me
	panic("implement me")
}

func (c *MeshCertProvider) ParseConfig(i interface{}) (*certprovider.BuildableConfig, error) {
	return certprovider.NewBuildableConfig("xx", nil, func(options certprovider.BuildOptions) certprovider.Provider {
		return &MeshCerts{}
	}), nil
}

func (c *MeshCertProvider) Name() string {
	return "mesh"
}

// XDSCreds provides credentials for authenticating with the XDS server.
// Token:
// - Istio-ca path
// - k8s token
// - MDS - if available
// - google default credentials
//
// Client certs:
// - workload id files
// - old istio files
//
// TransportCredentials also sets the expected CA and SAN for the server.
type XDSCreds struct {
}

func (x *XDSCreds) TransportCredentials() credentials.TransportCredentials {
	//TODO implement me
	return nil
}

func (x *XDSCreds) PerRPCCredentials() credentials.PerRPCCredentials {
	//TODO implement me
	return nil
}

func (x *XDSCreds) NewWithMode(mode string) (credentials.Bundle, error) {
	return nil, nil
}

func (x *XDSCreds) Build(config json.RawMessage) (credentials.Bundle, error) {
	return x, nil
}

func (x *XDSCreds) Name() string {
	return "mtls"
}
