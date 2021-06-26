package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/costinm/istiod/k8s"
)

func main() {
	gcpProj := os.Getenv("PROJECT")
	location := os.Getenv("LOCATION")
	cluster := os.Getenv("CLUSTER")

	// Used to provide access to all clusteres in the mesh
	meshID := os.Getenv("MESH_ID")
	meshIDLabel := os.Getenv("MESH_ID_LABEL")

	cfg := k8s.NewKubeConfig()

	if meshIDLabel == "" {
	  // if location is specified, create a single-cluster config.
		err := k8s.CreateClusterConfig(cfg, gcpProj, location, cluster)
		if err != nil {
			panic(err)
		}
	} else {
		k8s.AllHub(cfg, gcpProj, cluster, meshIDLabel, meshID)
		k8s.AllClusters(cfg, gcpProj, cluster, meshIDLabel, meshID)
	}
	err := k8s.SaveKubeConfig(cfg)
	if err != nil {
		panic(err)
	}
	tmpl := template.New("secret")
	tmpl, _ = tmpl.Parse(`
apiVersion: v1
kind: Secret
metadata:
  annotations:
    networking.istio.io/cluster: {{ name }}
  labels:
    istio/multiCluster: "true"
  name: istio-remote-secret-{{ name }}
  namespace: istio-system
stringData:
  {{ name }}: |
    apiVersion: v1
    kind: Config
    clusters:
    - cluster:
        certificate-authority-data: {{ ca }}
        server: {{ server }}
      name: {{ name }}
    contexts:
    - context:
        cluster: {{ name }}
        user: {{ name }} 
      name: {{ name }}
    current-context: {{ name }}
    preferences: {}
    users:
    - name: {{ name }}
---
`)
	for cn, _ := range cfg.Contexts {
		buf := &bytes.Buffer{}
		tmpl.Execute(buf, map[string]string{
			"name": cn,
			"ca": string(cfg.Clusters[cn].CertificateAuthorityData),
			"server": cfg.Clusters[cn].Server,
		})

		ioutil.WriteFile("secret-" + cn + ".yaml", buf.Bytes(), 0700)
	}
}

