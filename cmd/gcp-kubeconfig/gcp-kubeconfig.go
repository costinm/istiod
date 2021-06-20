package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"

	container "cloud.google.com/go/container/apiv1"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"k8s.io/client-go/rest"
)

var (
	gcpProj = flag.String("project", "", "GCP project")
	gcpZone = flag.String("location", "us-central1-c", "GCP zone")
	gkeCluster = flag.String("cluster", "big1", "GKE cluster")
)

func main() {
	flag.Parse()
	// TODO: if project/location not specified, get from local metadata server.
	// TODO: if cluster not specified, list clusters
	// TODO: use hub as well.

	cfg, err := clusterConf(*gcpProj, *gcpZone, *gkeCluster)
	if err != nil {
		panic(err)
	}
	cfgjs, err := json.Marshal(cfg)
	fmt.Println(string(cfgjs))
}

func clusterConf(p, l, clusterName string) (*rest.Config, error) {
	ctx := context.Background()

	cl, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return nil, err
	}

	c, err := cl.GetCluster(ctx, &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/cluster/%s", p, l, clusterName),
	})
	if err != nil {
		return nil, err
	}

	caCert, err := base64.StdEncoding.DecodeString(c.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return nil, err
	}
	cfg := &rest.Config{
		Host: "https://" + c.Endpoint,
		AuthProvider: &clientcmdapi.AuthProviderConfig{
			Name: "gcp",
		},
		TLSClientConfig: rest.TLSClientConfig{
			CAData: caCert,
		},
	}

	cfg.TLSClientConfig.CAData = caCert

	return cfg, nil
}
