package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/costinm/istiod/k8s"
	"github.com/costinm/istiod/ssh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)


func main() {
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		ns = "default"
	}

	k8sClient, err := k8s.GetK8S()
	if err != nil {
		panic(err)
	}

	for _, kv := range os.Environ() {
		kvl := strings.SplitN(kv, "=", 2)
		if strings.HasPrefix(kvl[0], "K8S_SECRET_") {
			InitSecret(k8sClient, ns, kvl[0][11:], kvl[1])
		}
		if kvl[0] == "SSH_CA" {
			InitDebug(ns, kvl[1])
		}
	}
	RefreshTokens(k8sClient, ns)
	InitToken(k8sClient, ns, "api", "kubernetes.io")

	xdsAddr := os.Getenv("XDS_ADDR")
	proxyConfig := os.Getenv("PROXY_CONFIG")
	if xdsAddr != "" || proxyConfig != "" {
		startIstioAgent(ns, xdsAddr)
	}



	startApp()

	select{}
}

// Refresh the tokens
func RefreshTokens(k8sClient *kubernetes.Clientset, ns string) {
	for _, kv := range os.Environ() {
		kvl := strings.SplitN(kv, "=", 2)
		if strings.HasPrefix(kvl[0], "K8S_TOKEN_") {
			InitToken(k8sClient, ns, kvl[0][10:], kvl[1])
		}
	}
	InitToken(k8sClient, ns, "istio-ca", "./var/run/secrets/tokens/istio-token")
	InitToken(k8sClient, ns, "api", "./var/run/secrets/kubernetes.io/serviceaccount/token")

}

func InitToken(client *kubernetes.Clientset, ns string, audience string, s2 string) {

}

// startApp uses the reminder of the command line to exec an app, using K8S_UID as UID, if present.
func startApp() {
	if len(os.Args) == 1 {
		return
	}

	cmd := exec.Command(os.Args[1], os.Args[1:]...)

	if os.Getuid() == 0 {
		uid := os.Getenv("K8S_UID")
		if uid != "" {
			uidi, err := strconv.Atoi(uid)
			if err == nil {
				cmd.SysProcAttr = &syscall.SysProcAttr{}
				cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uidi)}
			}
		}
	}

	go cmd.Start()

	go func() {
		cmd.Wait()
		os.Exit(0)
	}()
}

func startIstioAgent(ns string, xdsAddr string) {
	env := os.Environ()
	env = append(env, "XDS_ROOT_CA=SYSTEM")
	env = append(env, "CA_ROOT_CA=SYSTEM")
	env = append(env, "JWT_POLICY=first-party-jwt")
	env = append(env, "ISTIO_META_DNS_CAPTURE=true")

	if os.Getuid() == 0 {
		cmd := exec.Command("/usr/local/bin/pilot-agent", "istio-iptables")
		cmd.Start()
		cmd.Wait()
	}

	proxyConfig := os.Getenv("PROXY_CONFIG")
	if proxyConfig == "" {
		proxyConfig = fmt.Sprintf(`{"discoveryAddress": "%s"}`, xdsAddr)
	}


	cmd := exec.Command("/usr/local/bin/pilot-agent", "proxy", "sidecar", "--domain", ns + ".svc.cluster.local")
	if os.Getuid() == 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{
			Uid: 1337,
			Gid: 1337,
		}
	}
	cmd.Env = env

	go cmd.Start()

	go func() {
		cmd.Wait()
		os.Exit(0)
	}()


	// TODO: wait for agent to be ready
}

func InitSecret(k8sClient *kubernetes.Clientset,  ns string, name string, path string) {
	s, err := k8sClient.CoreV1().Secrets(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
			panic(err)
		}
	for k, v := range s.Data {
		err = ioutil.WriteFile(path + "/" + k, v, 0700)
		if err != nil {
			log.Println("Failed to init secret ", name, path, k, err)
		}
	}
}


func InitDebug(ns string, sshca string) {
	err := ssh.StartSSHDWithCA(ns, sshca)
	if err != nil {
		panic(err)
	}
}
