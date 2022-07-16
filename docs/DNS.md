# Istio DNS notes

## Development/debug

Initial install (clean cluster):
```

helm3 template istio-base manifests/base/ |kubectl apply -f -

helm3 install -n istio-system istio-16  manifests/istio-control/istio-discovery \
  -f manifests/global.yaml


kubectl apply -f https://k8s.io/examples/admin/dns/dnsutils.yaml
kubectl apply -f https://github.com/costinm/istiod/test/k8s/dns-custom-se.yaml

```

Updates:

```

make push.docker.pilot

helm3 upgrade -n istio-system istio-16  manifests/istio-control/istio-discovery \
  -f manifests/global.yaml --set global.tag=16 --set global.hub=costinm


 kubectl exec -ti dnsutils -- dig @ -p 53 

 # Global, using "proxy DNS", coreDNS
 kubectl exec -ti dnsutils -- dig @istiod.istio-system.svc +noall +answer -p 15054 customdns.example.global.

 # .mesh, using "proxy grpc", coreDNS
 kubectl exec -ti dnsutils -- dig @istiod.istio-system.svc +noall +answer -p 15054 customdns.example.mesh.

 # External, coreDNS
 kubectl exec -ti dnsutils -- dig @istiod.istio-system.svc +noall +answer -p 15054 www.google.com
 
 # cluster.local, coreDNS
 kubectl exec -ti dnsutils -- dig @istiod.istio-system.svc +noall +answer -p 15054 istiod.istio-system.svc.cluster.local.
  
 # Global, direct DNS
 kubectl exec -ti dnsutils -- dig @istiod.istio-system.svc +noall +answer customdns.example.global.

 # External, direct DNS
 kubectl exec -ti dnsutils -- dig @istiod.istio-system.svc +noall +answer -p 15054 customdns.example.mesh.
 
 # cluster.local, direct DNS
 kubectl exec -ti dnsutils -- dig @istiod.istio-system.svc +noall +answer customdns.example.mesh.
   
```

## Kube DNS

https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/

- both coreDNS and kube-dns are named 'kube-dns'
- kubelet passes DNS to each container with --cluster-dns=DNS-service-ip
- Pod 'dnsPolicy' - default inherits the node resolution
- kubelet can specify different resolv.conf file

- default corefile - in configmap 'coredns'
- 

## CoreDNS

- supports all major protocols
- DNS must be in files
- 

## TODO

- P0:  run DNS on agent, in forward-only mode

- P1: run DNS on agent, using CDS+EDS - using gRPC 
XDS protocol if possible (side effect: support for gRPC in istiod)

- P2: expand the XDS in agent to include all resources 
- P2: expand XDS in istiod to include all resources

----
```
# Used when running CoreDNS along Istiod.
#
# Will be placed in /var/lib/istio/coredns - contains default values.
# For customization, mount a configmap in the directory.
#
# Ideally the K8S native CoreDNS can be configured to add DNS-over-TLS and
# the grpc call to istiod. Until we find a good way to do this we need a way
# to resolve recursive calls without an un-encrypted call to kube dns.
#
# WIP: deciding if DNS-over-TLS will be handled directly by Istiod or by the CoreDNS.
# Not clear if we need an option, or just testing for both alternatives.
#
# Will connect to the local istiod using gRPC
# TODO: instead of proxy, use the k8s config directly
# TODO: add tls://, https:// and grpc:// listeners with same config

# Based on default k8s config, with changes:
# - different ports - not running as root, Service translates
# - log - for debugging


```

# CoreDNS proxy

- one active message per connection.
- mike appears to do one connection per message - co.Dial/co.Close in exchange
    - read,write deadline, WriteMsg/ReadMsg
    - uses Conn type - which helds a net.Conn 
https://github.com/miekg/dns/issues/614 - difficult to implement and out of scope

dns.Conn: properly implements Write, in WriteMsg

https://github.com/kubernetes/kubernetes/issues/73254 - use forward, which caches - instead of proxy

https://tools.ietf.org/html/rfc7858 "should not wait, pipeline"

# Local debug

mkdir /etc/istio
mkdir /var/lib/istio
chown costin /etc/istio/ /var/lib/istio

ln -s /ws/istio-stable/src/istio.io/istio/out/linux_amd64/envoy /usr/local/bin/

mkdir ../istiod/var/run/secrets/istio
