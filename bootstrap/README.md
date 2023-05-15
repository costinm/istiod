# XDS Bootstrap

- NAMESPACE
- Location

## GCP specific config

From MDS:

- PROJECT_ID
- PROJECT_NUMBER

From GKE MDS (only when running in GKE):

- CLUSTER_NAME
- CLUSTER_LOCATION

## MDS

- instanceID
- Tokens
- project id

## K8S JWT

```json
{
  "aud": [
    "https://container.googleapis.com/v1/projects/costin-asm1/locations/us-central1-c/clusters/big1"
  ],
  "exp": 1706276151,
  "iat": 1674740151,
  "iss": "https://container.googleapis.com/v1/projects/costin-asm1/locations/us-central1-c/clusters/big1",
  "kubernetes.io": {
    "namespace": "fortio-asm",
    "pod": {
      "name": "fortio-7b8dd44578-m8l5g",
      "uid": "b32a3b54-31c9-429c-bddf-25fbe9960a96"
    },
    "serviceaccount": {
      "name": "default",
      "uid": "3f5d5c4f-0e16-4c0c-9339-3df707d47e2c"
    },
    "warnafter": 1674743758
  },
  "nbf": 1674740151,
  "sub": "system:serviceaccount:fortio-asm:default"
}
```
