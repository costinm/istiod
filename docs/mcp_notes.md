
# Client

- Client code: pkg/mcp/sink/client_sink.go Run() - retry, backoff, wrapper around gRPC "EstablishResourceStream"
- pkg/mcp/sink/sink.go -> Recv and handleResponse

# Server

- pkg/mcp/source/server_source.go
- pkg/mcp/source/source.go 
- uses a con.queue
- eventually calls snapshot.Watch, with a queue
- 

K8S: galley/pkg/config/source/kube/apiserver/start.go

