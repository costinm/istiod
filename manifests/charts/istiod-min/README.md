WIP - not working yet.

Minimal Istiod install, with no Istio-specific options.

This chart has basic k8s pod/scale to create an Istiod with built-in default config.

Injection templates, mesh config and all other configs should be dynamic, and may
be added with kubectl or other charts - for example istiod-config - without touching
istiod install.


