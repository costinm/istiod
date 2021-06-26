//go:generate  protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_opt=Mssh-signer.proto=github.com/costinm/istiod/ssh --go_out=. ssh-signer.proto

package ssh

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	gossh "golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	xdscreds "google.golang.org/grpc/credentials/xds"
)

func StartSSHDWithCA(ns string, sshCA string) error {
	creds := insecure.NewCredentials()

	xdsBootstrap := os.Getenv("GRPC_XDS_BOOTSTRAP")
	if xdsBootstrap != "" {
			log.Println("Using xDS credentials...")
			var err error
			if creds, err = xdscreds.NewClientCredentials(xdscreds.ClientOptions{FallbackCreds: insecure.NewCredentials()}); err != nil {
				log.Fatalf("failed to create client-side xDS credentials: %v", err)
			}
	}

	conn, err := grpc.Dial("[::1]:8080", grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c := NewSSHCertificateServiceClient(conn)

	privk1, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	casigner1, _ := gossh.NewSignerFromKey(privk1)
	//pk := privk1.Public().(*ecdsa.PublicKey)
	//pkb := elliptic.Marshal(elliptic.P256(), pk.X, pk.Y)
	//pubk := base64.StdEncoding.EncodeToString(pkb)
	req := &SSHCertificateRequest{
		Public: string(gossh.MarshalAuthorizedKey(casigner1.PublicKey())),
	}
	log.Println(req)
	r, err := c.CreateCertificate(ctx, req)
	if err != nil {
		log.Println("Error creating cert ", err)
		return err
	}

	key, _, _, _, _ := gossh.ParseAuthorizedKey([]byte(r.Host))
	cert, ok := key.(*gossh.Certificate)

	fmt.Println(r.User)

	if !ok {
		return errors.New("unexpected cert")
	}
	signer, _ := gossh.NewCertSigner(cert, casigner1)

	ssht, err := NewSSHTransport(signer, "", ns, r.Root)
	if err != nil {
		return err
	}
	go ssht.Start()
	return nil
}


type SSHSigner struct {
	UnimplementedSSHCertificateServiceServer
	Root string
	Signer gossh.Signer
}

func (s *SSHSigner) CreateCertificate(ctx context.Context, in *SSHCertificateRequest) (*SSHCertificateResponse, error) {
	// TODO: get identity from JWT or cert or metadata

	pub, _, _, _, err := gossh.ParseAuthorizedKey([]byte(in.Public))
	if err != nil {
		return nil, err
	}
	log.Println("Creating certificate for ", in.Public)
	return &SSHCertificateResponse{
		Host: string(s.SignHost(pub, "example.com")),
		User: string(s.SignUser(pub, "user@example.com")),
		Root: s.Root,
	}, nil
}

func (s *SSHSigner) SignHost(pub gossh.PublicKey, name string) []byte {

	cert := &gossh.Certificate{
		ValidPrincipals: []string{name},
		Key:             pub,
		ValidBefore:     gossh.CertTimeInfinity,
		CertType:        gossh.HostCert,
	}
	cert.SignCert(rand.Reader, s.Signer)

	return gossh.MarshalAuthorizedKey(cert)
}

func (s *SSHSigner) SignUser(pub gossh.PublicKey, name string) []byte {

	cert := &gossh.Certificate{
		ValidPrincipals: []string{name},
		Key:             pub,
		ValidBefore:     gossh.CertTimeInfinity,
		CertType:        gossh.UserCert,
	}
	cert.SignCert(rand.Reader, s.Signer)

	return gossh.MarshalAuthorizedKey(cert)
}
