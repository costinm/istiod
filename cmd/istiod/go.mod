module github.com/costinm/istiod/cmd/istiod

go 1.16

// Cut&paste from istio go.mod

replace github.com/spf13/viper => github.com/istio/viper v1.3.3-0.20190515210538-2789fed3109c

// Old version had no license
replace github.com/chzyer/logex => github.com/chzyer/logex v1.1.11-0.20170329064859-445be9e134b2

// Avoid pulling in incompatible libraries
replace github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d

replace github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible

// Client-go does not handle different versions of mergo due to some breaking changes - use the matching version
replace github.com/imdario/mergo => github.com/imdario/mergo v0.3.5

// End cut&paste

//replace github.com/golang/glog => github.com/istio/glog v0.0.0-20190424172949-d7cfb6fa2ccd

//replace k8s.io/klog => github.com/istio/klog v0.0.0-20190424230111-fb7481ea8bcf

// See https://github.com/kubernetes/kubernetes/issues/92867, there is a bug in the library
//replace github.com/evanphx/json-patch => github.com/evanphx/json-patch v0.0.0-20190815234213-e83c0a1c26c8

//replace github.com/codahale/hdrhistogram => github.com/HdrHistogram/hdrhistogram-go v1.1.0
//
//replace github.com/docker/spdystream => github.com/moby/spdystream v0.2.0
//
//replace github.com/lyft/protoc-gen-validate => github.com/envoyproxy/protoc-gen-validate v0.6.1

//replace istio.io/istio => github.com/costinm/istio v0.0.0-20200727163637-9c8656454363

replace istio.io/istio => ../../../istio

require (
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	istio.io/istio v0.0.0-20210515053258-c8debb3e023c
	istio.io/pkg v0.0.0-20230411034200-2c98fd007de2
)
