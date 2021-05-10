module github.com/costinm/istiod

go 1.15

replace github.com/golang/glog => github.com/istio/glog v0.0.0-20190424172949-d7cfb6fa2ccd

replace k8s.io/klog => github.com/istio/klog v0.0.0-20190424230111-fb7481ea8bcf

replace github.com/spf13/viper => github.com/istio/viper v1.3.3-0.20190515210538-2789fed3109c

// See https://github.com/kubernetes/kubernetes/issues/92867, there is a bug in the library
replace github.com/evanphx/json-patch => github.com/evanphx/json-patch v0.0.0-20190815234213-e83c0a1c26c8

//replace istio.io/istio => github.com/costinm/istio v0.0.0-20200727163637-9c8656454363

replace istio.io/istio => /ws/istio-stable/src/istio.io/istio

require (
	github.com/envoyproxy/go-control-plane v0.9.9-0.20210115003313-31f9241a16e6
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.4.3
	istio.io/api v0.0.0-20210201080711-e51932d6679a
	istio.io/istio v0.0.0-20210201153422-44ad5ee0c4a1
	istio.io/pkg v0.0.0-20201230223204-2d0a1c8bd9e5
)
