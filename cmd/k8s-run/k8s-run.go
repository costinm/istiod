package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/costinm/cert-ssh/ssh"
	"github.com/creack/pty"

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
	name := os.Getenv("LABEL_APP")
	if name == "" {
		name = "default"
	}
	prefix := "."
	if os.Getuid() == 0 {
		prefix = ""
	}

	k8sClient, err := k8s.GetK8S()
	if err != nil {
		panic(err)
	}

	kr := &K8SRun{
		Name: name,
		Namespace: ns,
	}

	if len(os.Args) == 1 {
		// Default gateway label for now, we can customize with env variables.
		kr.Gateway = "ingressgateway"
	}

	for _, kv := range os.Environ() {
		kvl := strings.SplitN(kv, "=", 2)
		if strings.HasPrefix(kvl[0], "K8S_SECRET_") {
			kr.Secrets2Dirs[kvl[0][11:]] = prefix + kvl[1]
			InitSecret(k8sClient, ns, kvl[0][11:], prefix + kvl[1])
		}
		if kvl[0] == "SSH_CA" {
			kr.SSHCA = kvl[1]
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
		kr.StartIstioAgent(ns, xdsAddr, prefix)
	}

	if kr.Gateway == "" {
		startApp()
	}

	select{}
}

type K8SRun struct {
	// Secrets to 'mount'
	Secrets2Dirs map[string]string

	// Config maps to 'mount'
	CM2Dirs map[string]string

	// Audience to files
	Aud2Files map[string]string

	// If not empty, will run Istio-agent as a gateway with the given istio: label.
	Gateway string
	SSHCA   string

	Name string
	Namespace string
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

// StartIstioAgent creates the env and starts istio agent.
// If running as root, will also init iptables and change UID to 1337.
func (kr *K8SRun) StartIstioAgent(ns string, xdsAddr string, prefix string) {
	// /dev/stdout is rejected - it is a pipe.
	// https://github.com/envoyproxy/envoy/issues/8297#issuecomment-620659781

	env := os.Environ()
	// XDS and CA servers are using system certificates ( recommended ).
	// If using a private CA - add it's root to the docker images, everything will be consistent
	// and simpler !
	env = append(env, "XDS_ROOT_CA=SYSTEM")
	env = append(env, "CA_ROOT_CA=SYSTEM")

	// Save the istio certificates - for proxyless or app use.
	os.MkdirAll(prefix + "/var/run/secrets/istio.io", 0755)
	os.MkdirAll(prefix + "/etc/istio/pod", 0755)
	if os.Getuid() == 0 {
		os.Chown(prefix + "/var/run/secrets/istio.io", 1337, 1337)
		os.Chown(prefix + "/etc/istio/pod", 1337, 1337)
	}
	ioutil.WriteFile("/etc/istio/pod/labels", []byte(fmt.Sprintf(`version="v1"
security.istio.io/tlsMode="istio"
app="%s"
service.istio.io/canonical-name="%s"
`, kr.Name, kr.Name)), 0777)
	env = append(env, "OUTPUT_CERTS=" + prefix + "/var/run/secrets/istio.io/")

	// This would be used if a audience-less JWT was present - not possible with TokenRequest
	// TODO: add support for passing a long lived 1p JWT in a file, for local run
	//env = append(env, "JWT_POLICY=first-party-jwt")


	if os.Getuid() == 0 { // && kr.Gateway != "" {
		cmd := exec.Command("/usr/local/bin/pilot-agent", "istio-iptables")
		cmd.Env = env
		cmd.Dir = "/"
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
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

	env = append(env, "ISTIO_META_DNS_CAPTURE=true")

	proxyConfig := os.Getenv("PROXY_CONFIG")
	if proxyConfig == "" {
		proxyConfig = fmt.Sprintf(`{"discoveryAddress": "%s"}`, xdsAddr)
	}

	env = append(env, "PROXY_CONFIG=" + proxyConfig)

	var cmd *exec.Cmd
	if kr.Gateway != "" {
		ioutil.WriteFile("/etc/istio/pod/labels", []byte(`version=v1-cloudrun
security.istio.io/tlsMode="istio"
istio="ingressgateway"
`), 0777)
		cmd = exec.Command("/usr/local/bin/pilot-agent", "proxy", "router", "--domain", ns+".svc.cluster.local")
	} else {
		cmd = exec.Command("/usr/local/bin/pilot-agent", "proxy", "sidecar", "--domain", ns+".svc.cluster.local")
	}
	var stdout io.ReadCloser
	if os.Getuid() == 0 {
		os.MkdirAll("/etc/istio/proxy", 777)
		os.Chown("/etc/istio/proxy", 1337, 1337)

		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{
			Uid: 1337,
			Gid: 1337,
		}
		//cmd.SysProcAttr.Setsid = true
		//cmd.SysProcAttr.Setctty = true
		pty, tty, err := pty.Open()
		if err != nil {
			log.Println("Error opening pty ", err)
			stdout, _ = cmd.StdoutPipe()
			os.Stdout.Chown(1337, 1337)
		} else {
			cmd.Stdout = tty
			err = tty.Chown(1337, 1337)
			if err != nil {
				log.Println("Error chown ", tty.Name(), err)
			} else {
				log.Println("Opened pyy", tty.Name(), pty.Name())
			}
			stdout = pty
		}
		cmd.Dir = "/"
	} else {
		cmd.Stdout = os.Stdout
	}
	cmd.Env = env

	cmd.Stderr = os.Stderr
	os.MkdirAll(prefix + "/var/lib/istio/envoy/", 0700)
	go func() {
		err := cmd.Start()
		if err != nil {
			log.Println("Failed to start ", cmd, err)
		}
		go func() {
			io.Copy(os.Stdout, stdout)
		}()
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
		log.Println("Failed to start ssh", err)
	}
}
