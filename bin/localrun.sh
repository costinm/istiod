# From John

proxy-remotelocal-bootstrap () {
        sudo mkdir -p /var/run/secrets/tokens /var/run/secrets/istio /etc/istio/config
        echo '{"kind":"TokenRequest","apiVersion":"authentication.k8s.io/v1","spec":{"audiences":["istio-ca"], "expirationSeconds":2592000}}' | kubectl create --raw /api/v1/namespaces/${1:-default}/serviceaccounts/${2:-default}/token -f - | jq -j '.status.token' | sudo tee /var/run/secrets/tokens/istio-token

        kubectl -n istio-system get secret istio-ca-secret -ojsonpath='{.data.ca-cert\.pem}' | base64 -d | sudo tee /etc/certs/root-cert.pem

        mkdir -p /tmp/local-proxy-config
        sudo touch /etc/istio/config/mesh

        cp $GOPATH/src/istio.io/istio/tools/packaging/common/envoy_bootstrap.json /tmp/local-proxy-config/bootstrap.json
        cp $GOPATH/src/istio.io/istio/out/linux_amd64/envoy /tmp/local-proxy-config/envoy

        cat <<EOF | envsubst | sudo tee /var/lib/istio/envoy/cluster.env
PROXY_CONFIG="
binaryPath: /tmp/local-proxy-config/envoy
configPath: /tmp/local-proxy-config
proxyBootstrapTemplatePath: /tmp/local-proxy-config/bootstrap.json
discoveryAddress: localhost:15012
statusPort: 15020
terminationDrainDuration: 0s
tracing: {}
concurrency: 1
"
CA_ADDR=localhost:15012
ISTIO_INBOUND_PORTS=*
INBOUND_PORTS_EXCLUDE=15090,15021,15020
ISTIO_SERVICE_CIDR='*'
KUBE_VIRT_INTERFACES=docker0
ISTIO_META_DNS_CAPTURE='true'
IPTABLES_TRACE_LOGGING='true'
PROV_CERT=
EOF
# ISTIO_AGENT_FLAGS=--log_output_level=grpc:debug
    sudo chmod 777 /var/run/secrets/istio
    sudo chmod -R 777 /etc/istio
    kubectl -n istio-system get secret istio-ca-secret -ojsonpath='{.data.ca-cert\.pem}' | base64 -d > /var/run/secrets/istio/root-cert.pem
    sudo chmod -R 777 /tmp/local-proxy-config
}

sudo systemctl start istio
# if you run into su issues, modify /etc/security/access.conf and add `+:istio-proxy:ALL`
