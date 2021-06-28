module github.com/costinm/istiod/cmd/k8s-run

go 1.16

replace github.com/costinm/istiod/k8s => ../../k8s

require (
	github.com/costinm/cert-ssh/ssh v0.0.0-20210628224517-765c848d80b7 // indirect
	github.com/costinm/cert-ssh/sshca v0.0.0-20210628220432-a23b998ca61c // indirect
	//github.com/costinm/cert-ssh/ssh latest
	//github.com/costinm/cert-ssh/sshca latest
	github.com/costinm/istiod/k8s v0.0.0-00010101000000-000000000000
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2

)
