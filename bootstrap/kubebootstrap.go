package bootstrap

import (
	"io/ioutil"
	"os"

	"github.com/costinm/meshauth"
	"sigs.k8s.io/yaml"
)

// Load a kube config file for meshauth.
// Used to bootstrap auth.
func LoadKubeconfig() (*meshauth.KubeConfig, error) {
	kc := os.Getenv("KUBECONFIG")
	if kc == "" {
		kc = os.Getenv("HOME") + "/.kube/config"
	}
	kconf := &meshauth.KubeConfig{}

	var kcd []byte
	if kc != "" {
		if _, err := os.Stat(kc); err == nil {
			// Explicit kube config, using it.
			// 	"sigs.k8s.io/yaml"
			kcd, err = ioutil.ReadFile(kc)
			if err != nil {
				return nil, err
			}
			err := yaml.Unmarshal(kcd, kconf)
			if err != nil {
				return nil, err
			}

			return kconf, nil
		}
	}
	return nil, nil
}
