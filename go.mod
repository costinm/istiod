module github.com/costinm/istio-vm

go 1.12

replace github.com/golang/glog => github.com/istio/glog v0.0.0-20190424172949-d7cfb6fa2ccd

replace k8s.io/klog => github.com/istio/klog v0.0.0-20190424230111-fb7481ea8bcf

replace github.com/spf13/viper => github.com/istio/viper v1.3.3-0.20190515210538-2789fed3109c

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/NYTimes/gziphandler v1.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/go-ini/ini v1.33.0 // indirect
	github.com/gogo/protobuf v1.3.0
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/prometheus/client_golang v0.9.3
	go.opencensus.io v0.21.0
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	google.golang.org/grpc v1.24.0
	gopkg.in/ini.v1 v1.42.0 // indirect
	istio.io/api v0.0.0-20190930180221-6b9215326786
	istio.io/istio v0.0.0-20191002224728-c1df5928d28c
	istio.io/pkg v0.0.0-20190905225920-6d0bbfe3b229

	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)

// Kubernetes makes it challenging to depend on their libraries. To get around this, we need to force
// the sha to use. All of these are pinned to the tag "kubernetes-1.15"
replace k8s.io/kubernetes => k8s.io/kubernetes v1.15.0

replace k8s.io/api => k8s.io/api v0.0.0-20190620084959-7cf5895f2711

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190620090043-8301c0bda1f0

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab

replace k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190620085212-47dc9a115b18

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190620085554-14e95df34f1f

replace k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190620085838-f1cb295a73c9

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190620085706-2090e6d8f84c

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190620085942-b7f18460b210

replace k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190620085912-4acac5405ec6

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190620085809-589f994ddf7f

replace k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190531030430-6117653b35f1

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190620090116-299a7b270edc

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190620090156-2138f2c9de18

replace k8s.io/component-base => k8s.io/component-base v0.0.0-20190620085130-185d68e6e6ea

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190620090013-c9a0fc045dc1

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20190620085625-3b22d835f165

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190620085408-1aef9010884e

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190620085325-f29e2b4a4f84

replace k8s.io/kubectl => k8s.io/kubectl v0.0.0-20190602132728-7075c07e78bf
