Custom/sample alternative charts for Istio:

- istio-webhook-tag is the helm equivalent of 'istioctl x tag' commands, allows creation
  of custom tags for Istio automatic injection. Manipulates MutatingWebHookConfigs, indepdent 
  of Istiod installs.
  
- gate installs only the gateway deployment - no service or other configs. You can create multiple
 installs, for traffic shifting. This is based on injection - most 'legacy' cleaned up. It is 
  intended as an example you can customize - but can also be used directly.

- gateway-service manages the Service. This is version-independent, can be modified
  without reinstalling the gateway and will be used by multiple deployments of the gateway.

- WIP: istiod-min is a minimal (deployment and related) istiod install, with default settings.

- WIP: istiod-config contains dynamic configs for istiod. In future, they can apply to multiple 
  deployments (revisions) if istiod. 

- WIP: istiod is a copy of the istiod chart, for convenience. Small changes will be made to facilitate
  using istiod-config ( and merged upstream ). Ignore
