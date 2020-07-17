Manually configured files to test a minimal pilot+ingress.

Will create 'micro-ingress' namespace and  support Ingress with annotation "kubernetes.io/ingress.class: istio-micro-ingress"

Still requires Citadel in istio-system ( 1.0 or later ), and the CRDs, but nothing else.



```
# On the node - connect to the node port
# This also works with 'host network'

# Alternative to curl
openssl s_client -showcerts   -connect localhost:32443 -servername fortio.example.com

# Debug
_kexec app=pilot discovery istio-micro-ingress bash

_kexec app=shell kubectl test-micro-ingress bash


istioctl -i istio-micro-ingress proxy-status --log_output_level debug

istioctl -i istio-micro-ingress proxy-config listeners -o json $(istioctl -i istio-micro-ingress proxy-status | grep istio-ingress | cut -d' ' -f 1)



# Get the Load balancer address ( if you can't use node address )
export ING=$(kubectl get -n istio-micro-ingress service istio-micro-ingress -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

 curl -v $ING/debug/configz -H "Host: pilot-micro-ingress.example.com"
  
```
