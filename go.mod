module github.com/costinm/istiod

go 1.16

// Avoid pulling in incompatible libraries
replace github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d

replace github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible

replace github.com/golang/glog => github.com/istio/glog v0.0.0-20190424172949-d7cfb6fa2ccd

replace k8s.io/klog => github.com/istio/klog v0.0.0-20190424230111-fb7481ea8bcf

replace github.com/spf13/viper => github.com/istio/viper v1.3.3-0.20190515210538-2789fed3109c

// See https://github.com/kubernetes/kubernetes/issues/92867, there is a bug in the library
replace github.com/evanphx/json-patch => github.com/evanphx/json-patch v0.0.0-20190815234213-e83c0a1c26c8

//replace istio.io/istio => github.com/costinm/istio v0.0.0-20200727163637-9c8656454363

//replace istio.io/istio => /ws/istio-stable/src/istio.io/istio

require (
	github.com/Azure/azure-sdk-for-go v42.3.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/envoyproxy/go-control-plane v0.9.9-0.20210408202003-cde9fa27f1d4
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.1
	github.com/google/ko v0.8.3 // indirect
	github.com/vdemeester/k8s-pkg-credentialprovider v1.17.4 // indirect
	gonum.org/v1/netlib v0.0.0-20190331212654-76723241ea4e // indirect
	istio.io/api v0.0.0-20210507141635-02def630fd33
	istio.io/istio v0.0.0-20210201153422-44ad5ee0c4a1
	istio.io/pkg v0.0.0-20210405163638-bd457cbec517
)
