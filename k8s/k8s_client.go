package k8s

import (
	"errors"
	"flag"
	"os"
	"strings"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"
)

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Init klog.InitFlags from an env (to avoid messing with the CLI of the app)
func init() {
	fs := &flag.FlagSet{}
	kf := strings.Split(os.Getenv("KLOG_FLAGS")," ")
	fs.Parse(kf)
	klog.InitFlags(fs)
}

// GetK8S gets the default k8s client, using environment variables to decide how.
//
func GetK8S() (*kubernetes.Clientset, error) {
	var clientset *kubernetes.Clientset

	// In cluster
	hostInClustser := os.Getenv("KUBERNETES_SERVICE_HOST")
	if hostInClustser != "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}
		return clientset, nil
	}

	// Explicit kube config - use it
	kc := os.Getenv("KUBECONFIG")
	if kc == "" {
		kc = os.Getenv("HOME") + "/.kube/config"
	}
	if _, err := os.Stat(kc); err == nil {
		config, err := clientcmd.BuildConfigFromFlags("", kc)
		if err != nil {
			return nil, err
		}
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		return clientset, nil
	}

	// LOCATION - get a GKE cluster
	gcpProj := os.Getenv("PROJECT")
	location := os.Getenv("LOCATION")
	cluster := os.Getenv("CLUSTER")
	if location != "" {
		rc, err := CreateRestConfig(gcpProj, location, cluster)
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(rc)
	}
	return nil, errors.New("not found")
}
