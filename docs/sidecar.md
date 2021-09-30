
# Using sidecar for local ports

## Inbound

- defaultEndpoint can be UDS or 127.0.0.1:port or 0.0.0.0:PORT (instance IP)
- bind: 

```yaml


workloadSelector:
  labels:
    app: fortio-cr
ingress:
  - port:
      number: 9999
      protocol: HTTP
      name: testsidecar
    defaultEndpoint: 127.0.0.1:8080

  - bind: 127.0.0.1
    captureMode: NONE 
    port:
      number: 7999
      protocol: HTTP
      name: testsidecar
    defaultEndpoint: 0.0.0.0:8080

    
```

Generates:

```json
    {
     "version_info": "2021-09-08T20:48:52Z/6",
     "cluster": {
      "@type": "type.googleapis.com/envoy.config.cluster.v3.Cluster",
      "name": "inbound|9999||",
      "type": "STATIC",
      "connect_timeout": "10s",
      "circuit_breakers": {
       "thresholds": [
        {
         "max_connections": 4294967295,
         "max_pending_requests": 4294967295,
         "max_requests": 4294967295,
         "max_retries": 4294967295,
         "track_remaining": true
        }
       ]
      },
      "metadata": {
       "filter_metadata": {
        "istio": {
         "services": []
        }
       }
      },
      "load_assignment": {
       "cluster_name": "inbound|9999||",
       "endpoints": [
        {
         "lb_endpoints": [
          {
           "endpoint": {
            "address": {
             "socket_address": {
              "address": "127.0.0.1",
              "port_value": 8080
             }
            }
           }
          }
         ]
        }
       ]
      }
     },
     "last_updated": "2021-09-08T20:49:14.404Z"
    },


```

If dest address is 0.0.0.0:
```json

    {
      
      "upstream_bind_config": {
       "source_address": {
        "address": "127.0.0.6",
        "port_value": 0
       }
      },
      "load_assignment": {
       "cluster_name": "inbound|7999||",
       "endpoints": [
        {
         "lb_endpoints": [
          {
           "endpoint": {
            "address": {
             "socket_address": {
              "address": "169.254.8.1",
              "port_value": 8080
             }
            }
           }
          }
         ]
        }
       ]
    },

```
