package k8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	container "cloud.google.com/go/container/apiv1"

	gkehub "cloud.google.com/go/gkehub/apiv1beta1"
	//gkehub "google.golang.org/genproto/googleapis/cloud/gkehub/v1beta1"
	gkehubpb "google.golang.org/genproto/googleapis/cloud/gkehub/v1beta1"

	crm "google.golang.org/api/cloudresourcemanager/v1"

	containerpb "google.golang.org/genproto/googleapis/container/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"k8s.io/client-go/rest"

	// Required for k8s client to link in the authenticator
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

)

// TODO: if project/location not specified, get from local metadata server.
// TODO: if cluster not specified, list clusters
// TODO: use hub as well.

// SaveKubeConfig saves the KUBECONFIG to ./var/run/.kube/config
// The assumption is that on a read-only image, /var/run will be
// writeable and not backed up.
func SaveKubeConfig(cfg *clientcmdapi.Config) error {
	cfgjs, err := clientcmd.Write(*cfg)
	if err != nil {
		return err
	}
	_ = os.Mkdir("./var", 0755)
	_ = os.Mkdir("./var/run", 0755)
	err = os.Mkdir("./var/run/.kube", 0755)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("./var/run/.kube/config", cfgjs, 0744)
	if err != nil {
		return err
	}
	return nil
}

func NewKubeConfig() *clientcmdapi.Config {
	return &clientcmdapi.Config{
		APIVersion: "v1",
		Contexts: map[string]*clientcmdapi.Context{
		},
		Clusters: map[string]*clientcmdapi.Cluster{
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
		},
	}
}

// CreateClusterConfig sets a single, default cluster
func CreateClusterConfig(cfg *clientcmdapi.Config, p, l, clusterName string) error {
	ctx := context.Background()

	cl, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return err
	}

	c, err := cl.GetCluster(ctx, &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/cluster/%s", p, l, clusterName),
	})
	if err != nil {
		return err
	}

	caCert, err := base64.StdEncoding.DecodeString(c.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return err
	}

	ctxName := "gke_" + p + "_" + l + "_" + clusterName

	// We need a KUBECONFIG - tools/clientcmd/api/Config object
	cfg.CurrentContext = ctxName
	cfg.Contexts[ctxName]= &clientcmdapi.Context {
				Cluster: ctxName,
				AuthInfo: ctxName,
	}
	cfg.Clusters[ctxName] = &clientcmdapi.Cluster{
				Server: "https://" + c.Endpoint,
				CertificateAuthorityData: caCert,
	}
	cfg.AuthInfos[ctxName] = &clientcmdapi.AuthInfo{
				AuthProvider: &clientcmdapi.AuthProviderConfig{
					Name: "gcp",
				},
	}

	return nil
}

func CreateRestConfig(p, l, clusterName string) (*rest.Config, error) {
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

	// This is a rest.Config - can be used directly with the rest API
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

func ProjectNumber(p string) string {
	ctx := context.Background()

	cr, err := crm.NewService(ctx)
	pdata, err := cr.Projects.Get(p).Do()
	if err != nil {
		log.Println("Error getting project number", p, err)
		return p
	}

	// This is in v1 - v3 has it encoded in name.
	return strconv.Itoa(int(pdata.ProjectNumber))
}

func AllHub(kcfg *clientcmdapi.Config, project string, defCluster string, label string, meshID string) error {
	ctx := context.Background()

	cl, err := gkehub.NewGkeHubMembershipClient(ctx)
	if err != nil {
		return err
	}

	//cl.GenerateConnectManifest()
	mi := cl.ListMemberships(ctx, &gkehubpb.ListMembershipsRequest{
		Parent: "projects/" + project + "/locations/-",
	})
	pn := ProjectNumber(project)
	for {
		r, err := mi.Next()
		//fmt.Println(r, err)
		if err != nil || r == nil {
			break
		}
		if label != "ALL" {
			if r.Labels[label] != meshID {
				continue
			}
		}

		mna := strings.Split(r.Name, "/")
		mn := mna[len(mna)-1]
		ctxName := "connectgateway_" + project + "_"  + mn
		kcfg.Contexts[ctxName] = &clientcmdapi.Context {
			Cluster:  ctxName,
			AuthInfo: ctxName,
		}
		kcfg.Clusters[ctxName] = &clientcmdapi.Cluster {
			Server: fmt.Sprintf("https://connectgateway.googleapis.com/v1beta1/projects/%s/memberships/%s",
				pn, mn),
		}
		kcfg.AuthInfos[ctxName] = &clientcmdapi.AuthInfo{
			AuthProvider: &clientcmdapi.AuthProviderConfig{
				Name: "gcp",
			},
		}

		if mn == defCluster {
			kcfg.CurrentContext = ctxName
		}

	}
	return nil
}

func AllClusters(kcfg *clientcmdapi.Config, project string, defCluster string, label string, meshID string) error {
	ctx := context.Background()

	cl, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return err
	}

	clusters, err := cl.ListClusters(ctx, &containerpb.ListClustersRequest{
		Parent: "projects/" + project + "/locations/-",
	})
	if err != nil {
		return err
	}

	for _, c := range clusters.Clusters {
		if label != "ALL" {
			if c.ResourceLabels[label] != meshID {
				continue
			}
		}

		caCert, err := base64.StdEncoding.DecodeString(c.MasterAuth.ClusterCaCertificate)
		if err != nil {
			return err
		}

		// This is a rest.Config - can be used directly with the rest API
		cfg := &rest.Config{
			Host: "https://" + c.Endpoint,
			AuthProvider: &clientcmdapi.AuthProviderConfig{
				Name: "gcp",
				Config: map[string]string{},
			},
			TLSClientConfig: rest.TLSClientConfig{
				CAData: caCert,
			},
		}

		cfg.TLSClientConfig.CAData = caCert

		ctxName := "gke_" + project + "_" + c.Location + "_" + c.Name

		// We need a KUBECONFIG - tools/clientcmd/api/Config object
		kcfg.Contexts[ctxName] = &clientcmdapi.Context {
					Cluster:  ctxName,
					AuthInfo: ctxName,
				}
		kcfg.Clusters[ctxName] = &clientcmdapi.Cluster {
					Server:                   "https://" + c.Endpoint,
					CertificateAuthorityData: caCert,
				}
		kcfg.AuthInfos[ctxName] = &clientcmdapi.AuthInfo{
			AuthProvider: &clientcmdapi.AuthProviderConfig{
				Name: "gcp",
			},
		}
		if c.Name == defCluster {
			kcfg.CurrentContext = ctxName
		}
	}
	return nil
}
