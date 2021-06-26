package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/costinm/istiod/ssh"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/zpages"
	gossh "golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/admin"
	"google.golang.org/grpc/credentials/insecure"
	xdscreds "google.golang.org/grpc/credentials/xds"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/xds"
	"k8s.io/client-go/kubernetes"
)

var k8sClient *kubernetes.Clientset

const rootca_dir = "./var/run/secrets/ssh-signer"
const rootca_file = "./var/run/secrets/ssh-signer/id_ecdsa"

func init() {
       if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
               log.Println("Failed to register ocgrpc server views: %v", err)
       }
       if err := view.Register(ocgrpc.DefaultClientViews...); err != nil {
               log.Println("Failed to register ocgrpc server views: %v", err)
       }
}

func main() {
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		ns = "istio-system"
	}

	var privk *ecdsa.PrivateKey
	var casigner gossh.Signer

	keyB, err := ioutil.ReadFile(rootca_file)
	if err == nil {
		casigner, err = gossh.ParsePrivateKey(keyB)
	}
	if err != nil || casigner == nil {
		log.Println("Failed to read key, generating", err)
		privk, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		casigner, _ = gossh.NewSignerFromKey(privk)
		ecb, _ := x509.MarshalECPrivateKey(privk)
		keyB := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: ecb})
		os.MkdirAll(rootca_dir, 0700)
		err = ioutil.WriteFile(rootca_file, keyB, 0700)
		if err != nil {
			log.Println("Failed to save private, using in-memory ", err)
		}
	} else {
	}

	// Didn't find a method to serialize the cert or the public key for authorized_key use
	//pkc := casigner.PublicKey().(gossh.CryptoPublicKey)
	//pk := pkc.CryptoPublicKey().(*ecdsa.PublicKey)
	//pkb := elliptic.Marshal(elliptic.P256(), pk.X, pk.Y)
	//pub64 := base64.StdEncoding.EncodeToString(pkb)
	authRoot := "cert-authority " + string(gossh.MarshalAuthorizedKey(casigner.PublicKey())) // ecdsa-sha2-nistp256 " + ssh.SSH_ECPREFIX + pub64 + " " + "root@ca"

	sshs := &ssh.SSHSigner{
		Signer: casigner,
		Root: authRoot,
	}


	servicePort := ":8080"
	greeterLis, err := net.Listen("tcp", servicePort)
	if err != nil {
		log.Fatalf("net.Listen(tcp, %q) failed: %v", servicePort, err)
	}

	creds := insecure.NewCredentials()

	grpcOptions := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}

	xdsBootstrap := os.Getenv("GRPC_XDS_BOOTSTRAP")
	if xdsBootstrap != "" {
		var err error
		if creds, err = xdscreds.NewServerCredentials(xdscreds.ServerOptions{FallbackCreds: insecure.NewCredentials()}); err != nil {
			log.Fatalf("failed to create server-side xDS credentials: %v", err)
		}
		grpcServer := xds.NewGRPCServer(grpcOptions...)
		ssh.RegisterSSHCertificateServiceServer(grpcServer, sshs)
		admin.Register(grpcServer)
		//reflection.Register(grpcServer)
		go grpcServer.Serve(greeterLis)
	} else {
		grpcServer := grpc.NewServer(grpcOptions...)
		ssh.RegisterSSHCertificateServiceServer(grpcServer, sshs)
		admin.Register(grpcServer)
		reflection.Register(grpcServer)

		go func () {
			err := grpcServer.Serve(greeterLis)
			if err != nil {
				panic(err)
			}
		}()
	}

	// Status
	mux := &http.ServeMux{}
	zpages.Handle(mux, "/debug")
	go http.ListenAndServe("127.0.0.1:8081", mux)

	// run-k8s helper can't start a debug ssh server if running ssh_signer -
	// no signer. Start one in-process, for debugging.
	err = ssh.StartSSHDWithCA(ns, "127.0.0.1" + servicePort)
	if err != nil {
		//panic(err)
	}

	select{}
}

