Directory includes a number of envoy configurations to be used for local testing.

It assumes /etc/certs on the machine is readable to the current user and includes some valid certificates.
istio/istio repo has some pre-generated certs. 

# envoy_local.json 

- from istio/istio testdata. 
- admin on :17000
- xds-grpc on 127.0.0.1:15010

Additional manual configs to local services, manually configured:
Listeners:
- :17003 TCP to 'pilot_http' on 127.0.0.1:17007
- :17001 TCP 'use_original_dst' - manual iptables port.
- :17002 HTTP proxy port, manual config
- :17011 TLS TCP to 127.0.0.1:17010
- :15090 HTTP to prometheus_stats cluster on 127.0.0.1:17000

