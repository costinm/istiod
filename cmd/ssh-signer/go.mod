module github.com/costinm/istiod/cmd/ssh-signer

go 1.16

replace github.com/costinm/istiod/ssh => ../../ssh

require (
	cloud.google.com/go v0.84.0 // indirect
	github.com/costinm/istiod/ssh v0.0.0-00010101000000-000000000000
	go.opencensus.io v0.23.0
	golang.org/x/crypto v0.0.0-20210503195802-e9a32991a82e
	google.golang.org/grpc v1.38.0
	k8s.io/client-go v0.21.2
)
