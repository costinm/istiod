module github.com/costinm/istiod

go 1.14

//replace istio.io/istio => /ws/istio-master/go/src/istio.io/istio

//replace istio.io/istio => github.com/costinm/istio v0.0.0-20191022214508-382391b5279c

// Reminder cut&pasted from istio/istio
replace github.com/golang/glog => github.com/istio/glog v0.0.0-20190424172949-d7cfb6fa2ccd

replace k8s.io/klog => github.com/istio/klog v0.0.0-20190424230111-fb7481ea8bcf

replace github.com/spf13/viper => github.com/istio/viper v1.3.3-0.20190515210538-2789fed3109c

replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20191011211953-adfac697dc5b

require (
	cloud.google.com/go v0.57.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.2.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/antlr/antlr4 v0.0.0-20200606165112-0b35a76e9b09 // indirect
	github.com/aokoli/goutils v1.0.1 // indirect
	github.com/armon/go-metrics v0.3.3 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cncf/udpa/go v0.0.0-20200508205342-3b31d022a144 // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible // indirect
	github.com/emicklei/go-restful v2.12.0+incompatible // indirect
	github.com/envoyproxy/go-control-plane v0.9.5 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.3.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/go-openapi/spec v0.19.8 // indirect
	github.com/go-openapi/swag v0.19.9 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/google/cel-go v0.5.1 // indirect
	github.com/google/go-cmp v0.4.1 // indirect
	github.com/google/go-github v17.0.0+incompatible // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/googleapis/gnostic v0.4.2 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/hashicorp/consul v1.7.3 // indirect
	github.com/hashicorp/go-hclog v0.14.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.2.0 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/miekg/dns v1.1.29 // indirect
	github.com/mitchellh/mapstructure v1.3.2 // indirect
	github.com/onsi/gomega v1.9.0
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/client_golang v1.6.0 // indirect
	github.com/prometheus/common v0.10.0 // indirect
	github.com/prometheus/procfs v0.1.0 // indirect
	github.com/prometheus/prom2json v1.3.0 // indirect
	github.com/prometheus/statsd_exporter v0.16.0 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0 // indirect
	github.com/yl2chen/cidranger v1.0.0 // indirect
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9 // indirect
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9 // indirect
	golang.org/x/sys v0.0.0-20200602225109-6fdc65e7d980 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	golang.org/x/tools v0.0.0-20200608174601-1b747fd94509 // indirect
	google.golang.org/genproto v0.0.0-20200608115520-7c474a2e3482 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	istio.io/api v0.0.0-20200606165740-d943a112f2f5
	istio.io/gogo-genproto v0.0.0-20200606170237-6733a0a86e7c // indirect
	istio.io/istio v0.0.0-20200608184557-c05e57fe55a3
	istio.io/pkg v0.0.0-20200606170016-70c5172b9cdf
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/cli-runtime v0.18.3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/helm v2.14.3+incompatible // indirect
	k8s.io/klog/v2 v2.1.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200427153329-656914f816f9 // indirect
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451 // indirect
	sigs.k8s.io/controller-runtime v0.6.0 // indirect
	sigs.k8s.io/service-apis v0.0.0-20200607184946-df6e630206a5 // indirect
	sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06 // indirect
)

replace github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.9.0

replace github.com/Azure/go-autorest/autorest/adal => github.com/Azure/go-autorest/autorest/adal v0.5.0

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.2.0+incompatible

replace github.com/istio.io/proxy/src/envoy/tcp/metadata_exchange/config => istio.io/istio/pilot/pkg/metadata_exchange v0.0.0-20200123191201-47e363a83438
