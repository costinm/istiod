# Configuration for istio-gate

This chart has configurations for "istio-gate" namespace. The deployment(s) are managed
separately, using commands like:

```shell
    # Install a gateway deployment using the canary version.
	helm  upgrade --install -n istio-gate \
			gate-canary \
			manifests/charts/gateway-workload \
     		  --set routerMode=sni-dnat \
    		--set revision=canary

```

You can have multiple deployments, each will be selected by the chart. 

## East-West gate

The gate config includes east-west configuration.

