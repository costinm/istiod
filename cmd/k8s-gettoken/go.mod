module k8s-gettoken

go 1.16

replace github.com/costinm/istiod/k8s => ../../k8s

require (
	github.com/costinm/istiod/k8s v0.0.0-00010101000000-000000000000
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/klog v1.0.0
)
