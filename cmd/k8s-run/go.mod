module github.com/costinm/istiod/cmd/k8s-run

go 1.16

replace github.com/costinm/istiod/k8s => ../../k8s

replace github.com/costinm/cert-ssh/ssh => ../../../cert-ssh/ssh

require (
	github.com/costinm/cert-ssh/ssh v0.0.0-20210628224517-765c848d80b7
	github.com/costinm/istiod/k8s v0.0.0-00010101000000-000000000000
	github.com/creack/pty v1.1.13
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2

)
