# East-West gateway config

Istio includes a sample script (./gen-eastwest-gateway.sh) that
takes network, cluster and mesh id as params and generates 
an operator config, which in turn is applied to create a 
gateway. 

This document explains how to achieve the same result with the 
split deployment/service model.

The relevant config is: 

```yaml
   ...
        k8s:
          env:
            # sni-dnat adds the clusters required for AUTO_PASSTHROUGH mode
            - name: ISTIO_META_ROUTER_MODE
              value: "sni-dnat"
            # traffic through this gateway should be routed inside the network
            - name: ISTIO_META_REQUESTED_NETWORK_VIEW
              value: ${NETWORK}
          service:
            ports:
              - name: status-port
                port: 15021
                targetPort: 15021
              - name: tls
                port: 15443
                targetPort: 15443
              - name: tls-istiod
                port: 15012
                targetPort: 15012
              - name: tls-webhook
                port: 15017
                targetPort: 15017
  values:
    global:
      meshID: ${MESH}
      network: ${NETWORK}
      multiCluster:
        clusterName: ${CLUSTER}

```

That means the service config will have the given ports, and 
the deployment will have:
```yaml
        - name: ISTIO_META_ROUTER_MODE
          value: sni-dnat
        - name: ISTIO_META_REQUESTED_NETWORK_VIEW
          value: ${NETWORK}
        - name: ISTIO_META_NETWORK
          value: ${NETWORK}
        - name: ISTIO_META_CLUSTER_ID
          value: ${CLUSTER}
        - name: ISTIO_META_MESH_ID
          value: ${MESH}

```

For gateway injection to work, we would need to configure istiod
with the extra setting in global. 

## Gateawy Config 

The gateway config consists of the 3 expose-{istiod,istiod-https,services} 
included in the sample gateway-config-all, and the additional ports in the 
service.yaml.
