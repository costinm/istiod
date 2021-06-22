package main

import (
	"os"

	"github.com/costinm/istiod/gcp"
)

func main() {
	gcpProj := os.Getenv("PROJECT")
	location := os.Getenv("LOCATION")
	cluster := os.Getenv("CLUSTER")
	meshID := os.Getenv("MESH_ID")
	meshIDLabel := os.Getenv("MESH_ID_LABEL")

	cfg := gcp.NewKubeConfig()

	if meshIDLabel == "" {
	  // if location is specified, create a single-cluster config.
		err := gcp.CreateClusterConfig(cfg, gcpProj, location, cluster)
		if err != nil {
			panic(err)
		}
	} else {
		gcp.AllHub(cfg, gcpProj, cluster, meshIDLabel, meshID)
		gcp.AllClusters(cfg, gcpProj, cluster, meshIDLabel, meshID)
	}
	err := gcp.SaveKubeConfig(cfg)
	if err != nil {
		panic(err)
	}
}

