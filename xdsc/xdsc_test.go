package xdsc

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/costinm/istiod/bootstrap"
	"github.com/costinm/meshauth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	// Required for k8s client to link in the authenticator
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// Get mesh config from the current cluster
func LoadConfig(ctx context.Context) (*meshauth.K8SCluster, error) {
	kconf, err := bootstrap.LoadKubeconfig()
	if err != nil {
		return nil, err
	}

	def, _, err := meshauth.InitK8S(ctx, kconf)
	if err != nil {
		return nil, err
	}

	return def, err
}

func TestMCP(t *testing.T) {

	ctx, cf := context.WithTimeout(context.Background(), 100*time.Second)
	defer cf()

	kr, err := LoadConfig(ctx)

	ma := meshauth.NewMeshAuth(nil)

	// The mesh configuration should have all the properties we need -
	// see the code for example content.

	m := map[string]interface{}{}
	m["CLUSTER_ID"] = kr.ClusterName
	m["NAMESPACE"] = kr.Namespace
	m["SERVICE_ACCOUNT"] = kr.ServiceAccount
	m["x-goog-user-project"] = kr.ProjectID
	// Required for connecting to MCP.
	//m["CLOUDRUN_ADDR"] = "asm-big1-asm-managed-us-central1-c-42okyzkgcq-uc.a.run.app:443"
	// kr.MeshTenant

	//m["LABELS"] = x.Labels

	m["ISTIO_VERSION"] = "1.20.0-xdsc"
	m["SDS"] = "true"

	sts1, err := kr.GCPFederatedSource(ctx)
	ts := meshauth.NewGCPTokenSource(&meshauth.GCPAuthConfig{
		TokenSource:   sts1,
		ProjectNumber: kr.GetEnv(meshauth.MeshProjectNumber, ""),
		TrustDomain:   kr.ProjectID + ".svc.id.goog",
	})

	xdsConfig := &Config{
		Namespace: kr.Namespace,
		Workload:  ma.MDS.WorkloadName() + "-" + "10.10.1.1",
		Meta:      m,
		NodeType:  "sidecar",
		IP:        "10.10.1.1",
		Context:   ctx,

		GrpcOpts: []grpc.DialOption{
			// Using the STS library to exchange tokens
			grpc.WithPerRPCCredentials(ts),
			grpc.WithTransportCredentials(credentials.NewTLS(
				&tls.Config{
					InsecureSkipVerify: false,
				})),
		},
	}
	ktok, err := kr.GetToken(ctx, kr.ProjectID+" .svc.id.goog")
	xdsConfig.XDSHeaders = map[string]string{
		"x-goog-user-project":  kr.ProjectID, // kr.GetEnv(meshauth.MeshProjectNumber, ""), //
		"x-mesh-authorization": ktok,
	}

	//ctx = metadata.AppendToOutgoingContext(context.Background(), "ClusterID", p.clusterID)
	//xdsAddr := kr.XDSAddr
	xdsAddr := //"staging-meshconfig.sandbox.googleapis.com:443"
		"meshconfig.googleapis.com:443"
	xdscc, err := DialContext(ctx, xdsAddr, xdsConfig)
	// calls Run()
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Connected", xdscc)
	xdscc.Watch()

	xdscc.Fetch()
	log.Println(xdscc.Responses)
}

func TestIstiod(t *testing.T) {

	ctx, cf := context.WithTimeout(context.Background(), 100*time.Second)
	defer cf()

	kr, err := LoadConfig(ctx)
	krts := kr.NewK8STokenSource("istio-ca")
	ma := meshauth.NewMeshAuth(nil)

	// The mesh configuration should have all the properties we need -
	// see the code for example content.

	m := map[string]interface{}{}
	m["CLUSTER_ID"] = kr.ClusterName
	m["NAMESPACE"] = kr.Namespace
	m["SERVICE_ACCOUNT"] = kr.ServiceAccount
	// kr.MeshTenant

	//m["LABELS"] = x.Labels

	m["ISTIO_VERSION"] = "1.20.0-xdsc"
	m["SDS"] = "true"

	xdsConfig := &Config{
		Namespace: kr.Namespace,
		Workload:  ma.MDS.WorkloadName() + "-" + "10.10.1.1",
		Meta:      m,
		NodeType:  "sidecar",
		IP:        "10.10.1.1",
		Context:   ctx,

		GrpcOpts: []grpc.DialOption{
			// Using the STS library to exchange tokens
			grpc.WithPerRPCCredentials(krts),
			grpc.WithTransportCredentials(credentials.NewTLS(
				&tls.Config{
					// TODO: specify the cert from kr
					InsecureSkipVerify: true,
				})),
		},
	}

	//ctx = metadata.AppendToOutgoingContext(context.Background(), "ClusterID", p.clusterID)
	xdsAddr := kr.GetEnv("MCON_ADDR", "") + ":15012"
	if xdsAddr == "" {
		// This is the external address, only enabled for multi-network.
		// If running from an internal address, use
		// kr.MeshConnectorInternalAddr
		xdsAddr = "127.0.0.1:15012"
	}

	xdscc, err := DialContext(ctx, xdsAddr, xdsConfig)
	// calls Run()
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Connected", xdscc)
	xdscc.Watch()

	xdscc.Fetch()
	log.Println(xdscc.Responses)
}

func TestTD(t *testing.T) {

	ctx, cf := context.WithTimeout(context.Background(), 100*time.Second)
	defer cf()

	kr, err := LoadConfig(ctx)

	m := map[string]interface{}{}
	m["TRAFFICDIRECTOR_NETWORK_NAME"] = "default"
	m["TRAFFICDIRECTOR_GCP_PROJECT_NUMBER"] = kr.GetEnv(meshauth.MeshProjectNumber, "")
	m["INSTANCE_IP"] = "10.48.0.63"

	sts1, err := kr.GCPFederatedSource(ctx)
	ts := meshauth.NewGCPTokenSource(&meshauth.GCPAuthConfig{
		TokenSource:   sts1,
		ProjectNumber: kr.GetEnv(meshauth.MeshProjectNumber, ""),
		TrustDomain:   kr.ProjectID + ".svc.id.goog",
	})
	//meshauth.NewGSATokenSource(&meshauth.STSAuthConfig{
	//	TokenSource: kr,
	//}, "k8s-fortio@costin-asm1.iam.gserviceaccount.com")),

	ma := meshauth.NewMeshAuth(nil)

	xdsConfig := &Config{
		Namespace: kr.Namespace,
		Workload:  ma.MDS.WorkloadName() + "-" + "10.10.1.1",
		Meta:      m,

		NodeType: "sidecar",
		IP:       "10.10.1.1",
		Context:  ctx,
		Locality: kr.Location,
		NodeId: fmt.Sprintf("projects/%s/networks/default/nodes/1234",
			kr.GetEnv(meshauth.MeshProjectNumber, "")),

		GrpcOpts: []grpc.DialOption{
			// Using the STS library to exchange tokens
			grpc.WithPerRPCCredentials(ts),
			grpc.WithTransportCredentials(credentials.NewTLS(
				&tls.Config{
					// TODO: specify the cert from kr
					InsecureSkipVerify: true,
				})),
		},
	}

	xdscc, err := DialContext(ctx, "trafficdirector.googleapis.com:443", xdsConfig)
	// calls Run()
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Connected", xdscc)
	xdscc.Watch()

	xdscc.Fetch()
	log.Println(xdscc.Responses)
}
