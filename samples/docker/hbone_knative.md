# HBONE for knative / CloudRun

Primary port is 15002, which will be the reserved port for HBONE over H2C. 

In H2C mode, HBONE will run behind a trusted load balancer that terminates TLS, and optionally mTLS, and 
forwards the request over H2C using POST method. This is intended for interoperability with external 
load balancers that terminate TLS.

The port is configured in bootstrap:

```yaml
{
  "name": "hbone-h2c",
  "address": {
    "socket_address": {
      "protocol": "TCP",
      "address": "0.0.0.0",
      "port_value": 15002
    }
  },
  "filter_chains": [
    {
      "filters": [
        {
          "name": "envoy.filters.network.http_connection_manager",
          "typed_config": {
            "@type": "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
            "stat_prefix": "hbone-h2c",
            "codec_type": "HTTP2",
            "http2_protocol_options": {
              "allow_connect": true
            },
            "upgrade_configs": [
              {
                "upgrade_type": "CONNECT",
                "connect_config": {
                  "allow_post": true
                }
              }
            ],
            "route_config": {
              "name": "hbone_route",
              "virtual_hosts": []
            },
            "http_filters": [
              {
                "name": "envoy.filters.http.router"
              }
            ],
          }
        }
      ]
    }
  ]
}


```

## Trusted - direct 

In trusted/direct mode, the LB is a trusted component - the TCP stream sent over HBONE goes directly to a local
port. For example, forwarding to a SSHD will generate a route like:


```json
  {
    "name": "local_service",
    "domains": [
      "*"
    ],
    "routes": [
      {
        "match": {
          "prefix": "/hbone/22",
          "headers": [
            {
              "name": ":method",
              "exact_match": "POST"
            }
          ]
        },
        "route": {
          "cluster": "cluster_1",
          "upgrade_configs": [
            {
              "upgrade_type": "CONNECT",
              "connect_config": {
                "allow_post": true
              }
            }
          ]
        }
      }
    ]
  }



```

## Legacy mTLS

In 'legacy mTLS' mode, the LB is not trusted, and we use the old inbound path (non-HBONE). 

The HBONE stream is directed to a local mTLS inbound port, with the typical generated config.

The path may be optimized with the internal redirect.
 
## Untrusted LB mode, HBONE*2

In this mode, the LB is used as an untrusted transport. The TCP stream is forwarded to the real HBONE-mTLS 
port, and behave identically with Istio in HBONE mode.
