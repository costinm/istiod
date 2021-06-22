module gcp-kubeconfig

go 1.16

replace github.com/costinm/istiod/gcp => ../../gcp

require (
	cloud.google.com/go v0.84.0
	github.com/costinm/istiod/gcp v0.0.0-00010101000000-000000000000
	google.golang.org/api v0.48.0
	google.golang.org/genproto v0.0.0-20210608205507-b6d2f5bf0d7d
	k8s.io/client-go v0.21.2
)
