# Ports used by Istio and Istio-on-a-VM


| Istio service | Port         | Description           |
| istio-pilot | 15080 (8080) | HTTP interface, debug |
| istio-pilot | 15010        | GRPC, plain           |
| istio-pilot | 15011        | GRPC, MTLS           |
| istio-pilot | 15014        | prometheus           |
| istio-pilot | 15876 (9876)  | ctrlz           |
| galley      | 15901        | GRPC, plain  |
| galley      | 15015        | monitoring  |
| galley      | 15877        | ctrlz  |

| citadel | 1 
