
Server implementations

- Pilot

- 

Clients

- ADSC

- test/framework/components/pilot/client.go
    -- insecure only
    -- used with NewPortForwarder, creating a proxy and newNative - direct
    -- doesn't appear to be used
    
- pilot/tools/debug/pilot_cli
    -- insecure
    -- portForwardPilot
    
- envoy/v2/init_test/connectADS - insecure, used by a lot of test
    -- returns the stream

- grpc - xds/internal/client/v2client.go
    -- Options: DialOptions, targetName, bootstrap.Config
    -- used by resolver/Builder
    -- xds_resolver_test - 
    -- sendRequest(stream, resourceNames, type, version)
    -- handleLDSResponse, etc.
    -- lds.go - get route from listener
