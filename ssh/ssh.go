package ssh

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

type Server struct {
	Port           int
	Shell          string
	AuthorizedKeys []ssh.PublicKey

	clientConfig *gossh.ClientConfig

	signer ssh.Signer

	// HandleConn can be used to overlay a SSH conn.
	server      *ssh.Server
	CertChecker *gossh.CertChecker
}

func NewSSHTransport(signer gossh.Signer, name, domain, root string) (*Server, error) {
	// privteKey crpto.Signer
	//signer, _ := gossh.NewSignerFromKey(privateKey)

	s := &Server{
		signer: signer,
		clientConfig: &gossh.ClientConfig{
			Auth: []gossh.AuthMethod{gossh.PublicKeys(signer)},
			HostKeyCallback: func(hostname string, remote net.Addr, key gossh.PublicKey) error {
				return nil
			},
			//Config: gossh.Config{
			//	MACs: []string{
			//		"hmac-sha2-256-etm@opengossh.com",
			//		"hmac-sha2-256",
			//		"hmac-sha1",
			//		"hmac-sha1-96",
			//	},
			//	Ciphers: []string{
			//		"aes128-gcm@opengossh.com",
			//		"chacha20-poly1305@opengossh.com",
			//		"aes128-ctr", "none",
			//	},
			//},
		},
		Port: 15022,
		Shell: "/bin/bash",
		CertChecker: &gossh.CertChecker{},
	}
	//pk, err := LoadAuthorizedKeys(os.Getenv("HOME") + "/.ssh/authorized_keys")
	//if err == nil {
	//	s.AuthorizedKeys = pk
	//}
	extra := os.Getenv("AUTHORIZED")
	if extra != "" {
		pubk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(extra))
		if err == nil {
			s.AuthorizedKeys = append(s.AuthorizedKeys, pubk)
		}
	}
	pubk, _,_, _, err := ssh.ParseAuthorizedKey([]byte(root))
	if err == nil {
		s.AuthorizedKeys = append(s.AuthorizedKeys, pubk)
	} else {
		log.Println(err)
		return nil, err
	}

	s.server = s.getServer(signer)

	return s, nil
}

func (srv *Server) getServer(signer ssh.Signer) *ssh.Server {
	forwardHandler := &ssh.ForwardedTCPHandler{}

	server := &ssh.Server{
		Addr:    fmt.Sprintf(":%d", srv.Port),
		Handler: srv.connectionHandler,
		ChannelHandlers: map[string]ssh.ChannelHandler{
			"direct-tcpip": ssh.DirectTCPIPHandler,
			"session":      ssh.DefaultSessionHandler,
		},
		LocalPortForwardingCallback: ssh.LocalPortForwardingCallback(func(ctx ssh.Context, dhost string, dport uint32) bool {
			log.Println("Accepted forward", dhost, dport)
			return true
		}),
		ReversePortForwardingCallback: ssh.ReversePortForwardingCallback(func(ctx ssh.Context, host string, port uint32) bool {
			log.Println("attempt to bind", host, port, "granted")
			return true
		}),
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
		},
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": sftpHandler,
		},
	}

	server.AddHostKey(signer)

	// It seems gliderlabs creates more problems than it solves...
	// Must be set, otherwise no auth happens.
	server.PublicKeyHandler = srv.authorize

	return server
}

func (t *Server) Start() {
	go t.server.ListenAndServe()
}







