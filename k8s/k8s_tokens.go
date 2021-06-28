package k8s

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type K8STokens struct {
	Namespace string
	KSA string
	Client *kubernetes.Clientset
	AudToFile map[string]string
}

func (kt *K8STokens) Refresh() {
	for aud, f := range kt.AudToFile {
		InitToken(kt.Client, kt.Namespace, kt.KSA, aud, f)
	}

	time.AfterFunc(30 * time.Minute, kt.Refresh)
}

func InitToken(client *kubernetes.Clientset, ns string, ksa string, audience string, s2 string) error {
	treq := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			Audiences: []string{audience},
		},
	}
	ts, err := client.CoreV1().ServiceAccounts(ns).CreateToken(context.Background(),
		ksa, treq, metav1.CreateOptions{})
	if err != nil {
		log.Println("Error creating ", ns, ksa, audience, err)
		return err
	}

	lastSlash := strings.LastIndex(s2, "/")
	err = os.MkdirAll(s2[:lastSlash], 0700)
	if err != nil {
		log.Println("Error creating dir", ns, ksa, s2[:lastSlash])
	}
	err = ioutil.WriteFile(s2, []byte(ts.Status.Token), 0700)
	if err != nil {
		log.Println("Error creating ", ns, ksa, audience, s2, err)
		return err
	}

	return nil
}
