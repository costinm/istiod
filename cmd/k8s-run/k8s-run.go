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

	"github.com/costinm/cert-ssh/ssh"

	"github.com/costinm/istiod/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)


func main() {
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		ns = "default"
	}
	ksa := os.Getenv("SERVICE_ACCOUNT")
	if ksa == "" {
		ksa = "default"
	}
	prefix := "."
	if os.Getuid() == 0 {
		prefix = ""
	}

	k8sClient, err := k8s.GetK8S()
	if err != nil {
		panic(err)
	}

	for _, kv := range os.Environ() {
		kvl := strings.SplitN(kv, "=", 2)
		if strings.HasPrefix(kvl[0], "K8S_SECRET_") {
			InitSecret(k8sClient, ns, kvl[0][11:], prefix + kvl[1])
		}
		if kvl[0] == "SSH_CA" {
			InitDebug(ns, kvl[1])
		}
	}
	RefreshTokens(k8sClient, ns, ksa, prefix)

	xdsAddr := os.Getenv("XDS_ADDR")
	if xdsAddr == "" {
		xdsAddr = "istiod.svc.i.webinf.info:443"
	}
	proxyConfig := os.Getenv("PROXY_CONFIG")
	if xdsAddr != "" || proxyConfig != "" {
		startIstioAgent(ns, xdsAddr, prefix)
	}

	startApp()

	select{}
}

// Refresh the tokens
func RefreshTokens(k8sClient *kubernetes.Clientset, ns, ksa, prefix string) {
	kt := k8s.K8STokens{
		KSA: ksa,
		Namespace: ns,
		Client: k8sClient,
		AudToFile: map[string]string{},
	}

	for _, kv := range os.Environ() {
		kvl := strings.SplitN(kv, "=", 2)
		if strings.HasPrefix(kvl[0], "K8S_TOKEN_") {
			kt.AudToFile[kvl[0][10:]] =  prefix + kvl[1]
		}
	}
	kt.AudToFile["istio-ca"] = prefix + "/var/run/secrets/tokens/istio-token"
	kt.AudToFile["api"] = prefix + "/var/run/secrets/kubernetes.io/serviceaccount/token"

	kt.Refresh()
}


// startApp uses the reminder of the command line to exec an app, using K8S_UID as UID, if present.
func startApp() {
	if len(os.Args) == 1 {
		return
	}

	var cmd *exec.Cmd
	if len(os.Args) == 2 {
		cmd = exec.Command(os.Args[1])
	} else {
		cmd = exec.Command(os.Args[1], os.Args[2:]...)
	}
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
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	go func() {
		err := cmd.Start()
		if err != nil {
			log.Println("Failed to start ", cmd, err)
		}
		err = cmd.Wait()
		if err != nil {
			log.Println("Failed to wait ", cmd, err)
		}
		os.Exit(0)
	}()
}

func startIstioAgent(ns string, xdsAddr string, prefix string) {
	env := os.Environ()
	env = append(env, "XDS_ROOT_CA=SYSTEM")
	env = append(env, "CA_ROOT_CA=SYSTEM")

	// This would be used if a audience-less JWT was present - not possible with TokenRequest
	//env = append(env, "JWT_POLICY=first-party-jwt")

	env = append(env, "ISTIO_META_DNS_CAPTURE=true")

	if os.Getuid() == 0 {
		cmd := exec.Command("/usr/local/bin/pilot-agent", "istio-iptables")
		cmd.Env = env
		cmd.Dir = "/"
		err := cmd.Start()
		if err != nil {
			log.Println("Error starting iptables", err)
		} else {
			err = cmd.Wait()
			if err != nil {
				log.Println("Error starting iptables", err)
			}
		}
	}

	proxyConfig := os.Getenv("PROXY_CONFIG")
	if proxyConfig == "" {
		proxyConfig = fmt.Sprintf(`{"discoveryAddress": "%s"}`, xdsAddr)
	}

	env = append(env, "PROXY_CONFIG=" + proxyConfig)

	cmd := exec.Command("/usr/local/bin/pilot-agent", "proxy", "sidecar", "--domain", ns + ".svc.cluster.local")
	if os.Getuid() == 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{
			Uid: 1337,
			Gid: 1337,
		}
		cmd.Dir = "/"
	}
	cmd.Env = env

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	os.MkdirAll(prefix + "/var/lib/istio/envoy/", 0700)
	go func() {
		err := cmd.Start()
		if err != nil {
			log.Println("Failed to start ", cmd, err)
		}
		err = cmd.Wait()
		if err != nil {
			log.Println("Wait err ", err)
		}

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
