module github.com/costinm/istiod/cmd/k8s-run

go 1.16

replace github.com/costinm/istiod/k8s => ../../k8s

replace github.com/costinm/istiod/ssh => ../../ssh

require (
	github.com/costinm/istiod/k8s v0.0.0-00010101000000-000000000000
	github.com/costinm/istiod/ssh v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20210503195802-e9a32991a82e
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
)
