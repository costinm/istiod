package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	ns = flag.String("n", "default", "Namespace")
	sa = flag.String("sa", "default", "Service account")
	aud = flag.String("aud", "istio-ca", "Audience. 'api' for k8s server")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	var err error

	treq := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			Audiences: []string{*aud},
		},
	}
	clientset, err := getK8S()
	if err != nil {
		panic(err)
	}
	ts, err := clientset.CoreV1().ServiceAccounts(*ns).CreateToken(context.Background(),
		*sa, treq, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println(ts.Status.Token)
}


func getK8S() (*kubernetes.Clientset, error) {
	var clientset *kubernetes.Clientset
	var err error
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
	kc := os.Getenv("KUBECONFIG")
	if kc == "" {
		kc = os.Getenv("HOME") + "/.kube/config"
	}
	config, err := clientcmd.BuildConfigFromFlags("", kc)
	if err != nil {
		return nil, err
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientset, nil
}

