# TODO

## Agent 

[] If custom root ca not found, use system. Remove the option - env is enough. (both xds and ca)
[] Same for JWT - use it from known locations, fallback to metadata server id token, google exchange if address has googleapis
[] attempt to derive the XDS address from the domain in the ID token. SA@domain.com -> istiod.domain.com


## Ideas

[] Combined Helm + Kustomize (as post-processing)
[] Auto-copy helm templates to istio/installer, generate helm install
[] Minimal istio-agent, no CLI. Pass all args to forked app
[] Generate grpc manifest automatically, use it as default

