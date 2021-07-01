Istio apps and gateway running in a single docker container.

This was originally (long, long time ago) intended for testing iptables
and VM. It is based on Istio for VM support.

Primary target is the base docker image, containing the installed proxy and a modified
startup script. This can be used as a base, adding your application.

# CloudRun

The docker images also work in CloudRun - or similar docker-based execution environments.

The images use HBONE to expose TCP ports and support e2e mTLS. 

## Inbound Ports 

The application is expected to use port 8080. We can add customization for the default port later - this matches
current default and practice in CloudRun/KNative. 

The container will listen on port 15009, as H2C and forward /_hbone/default (as an opaque stream)
to port 8443.

Port 8443 will have a real listener, configured in bootstrap and accepting mTLS and 
forwarding to port 8080. In future the listener on 15009 may do an internal forward.

Since HBONE on 15009 originates the connection - we can't rely on the inbound chain, 
the request is not intercepted. The config for port 8443 is a manual configuration for the
'mTLS listener' that Istio normally generates.



# Limitations/missing

- ./etc/istio/pod/annotations, labels
- ...
