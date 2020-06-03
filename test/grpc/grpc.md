Native GRPC support 

Env variable: GRPC_XDS_BOOTSTRAP

Istio requires at least a node ID, in the expected format: sidecar~IP~name.ns~ns.cluster.local

GRPC sends LDS requests with the "resource name" set to match the gRPC target.
This is one way to identify gRPC clients. Envoy sends *.

The node metadata also includes "gRPC' in UserAgentName and BuildVersion.
A "ClientFeatures" is present - envoy.lb.does_not_support_overprovisioning included.

It expects an "ApiListener" field inside the Listener response - containing  
httppb.HttpConnectionManager as "any". 

Currently in-line doesn't seem to be supported - instead RDS.routeConfigName is used with 
RDS to get the routes.

# DNS

For interception, we will use the same mechanism to determine all 'names' and VIPs:

Listener {
 name: "productpage.bookinfo_port"
 address: VIP:port
 api_listener: {
   {
     "route_config": {
        "name": "xxx"
     }
   }
 }
}

No need for RDS and extra RT. 

In normal Istio we have FilterChains, including one that contains the HttpConnectionManager.

Listener also includes metadata - not used for now. 

trafficDirection: Outbound

"no other field except for name should be set" - but address is mandatory.

# gRPC

To support gRPC we need to detect it's a grpc client, using the name, and return 
a similar response. 

"Resolver" - from Dial address, return the 'address' (VIP or list of endpoints), ServerName, other properties of the service.

"Balancer" - SubConn.UpdateAddresses(resolver.Address), SubConn.Connect

# Issues

- CDS - NACK/error if STATIC cluster is present in response, no reason for that
- RDS - only checks "", not "/"
- General: can skip all LDS/RDS hoopla, go straight to CDS or even EDS, using name:port
- RDS - should actually check the routes for the service (not only hostname - service and even method should be matched)

Resolver: ServerName should also support spiffee identity
