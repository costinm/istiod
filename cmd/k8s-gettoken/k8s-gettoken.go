package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	k8s "github.com/costinm/istiod/k8s"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	aud = flag.String("aud", "istio-ca", "Audience. 'api' for k8s server")
)

func main() {
	flag.Parse()

	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		ns = "default"
	}
	ksa := os.Getenv("POD_SERVICE_ACCOUNT")
	if ksa == "" {
		ksa = "default"
	}

	clientset, err := k8s.GetK8S()
	if err != nil {
		panic(err)
	}
	treq := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			Audiences: []string{*aud},
		},
	}
	if err != nil {
		panic(err)
	}
	ts, err := clientset.CoreV1().ServiceAccounts(ns).CreateToken(context.Background(),
		ksa, treq, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println(ts.Status.Token)
}


