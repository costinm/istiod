```shell
istioctl x revision tag generate stable2 -r v110 --skip-confirmation
```

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  creationTimestamp: null
  labels:
    app: sidecar-injector
    install.operator.istio.io/owning-resource: unknown
    istio.io/rev: v110
    istio.io/tag: stable2
    operator.istio.io/component: Pilot
    release: istio
  name: istio-revision-tag-stable2
webhooks:
- admissionReviewVersions:
  - v1beta1
  - v1
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMzakNDQWNhZ0F3SUJBZ0lSQU5GYmZObklMejhyNy9xdytidFNVazh3RFFZSktvWklodmNOQVFFTEJRQXcKR0RFV01CUUdBMVVFQ2hNTlkyeDFjM1JsY2k1c2IyTmhiREFlRncweU1EQTJNVGN3TVRVM01qRmFGdzB6TURBMgpNVFV3TVRVM01qRmFNQmd4RmpBVUJnTlZCQW9URFdOc2RYTjBaWEl1Ykc5allXd3dnZ0VpTUEwR0NTcUdTSWIzCkRRRUJBUVVBQTRJQkR3QXdnZ0VLQW9JQkFRQ3dmZnNKVkFDbnI1cjBSMHArcWdBRVFhWlVXMFR2M0dqTXkyYXAKelN2aEl3THFFSXNaWjE2WkxVZU1zaDEzYUFXTmRObDVkWkhZTmFZMHVJVzZsZlJ0bHJ6T2cwL295L3BJM2tBRgpwWEc2ajlXWkhmMVFVWkpHWnRFTFpXZDU1dmdwcTNtSHZ6SmhPZk0xT0RwMENnYkg0SVoxaHBsdnhPdnNFYkFxCjZ6Y21TMzNHc1ZkU2VKNjIwWGdYL2tBejlFRTFqVlVqUDV5eUJwYXlnQ3VVdHFCZ1JzbGhleWNkeEFvMTZhMVEKSTV3enZhTkdYa3IrQjhaaXFPUi9DWU5vdUhHbERSdk9JM25wY25JMnFkY2hKTkcwczA0S3ZJL0xNWWd4UytBbgpMdFp5VmJ2d1N3a1hkN2lxNjdPYlkrTkUxc3c2NGJZVEY2NUVwK0xtYnpITkgvVC9BZ01CQUFHakl6QWhNQTRHCkExVWREd0VCL3dRRUF3SUNCREFQQmdOVkhSTUJBZjhFQlRBREFRSC9NQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUIKQVFCT0NuRFV2RE4yRFEzWW9ILzdya3B1emF3SFYwUEE0aVhKUnhjSFZVTlpleXRtRUsyWGdVV2JHYURpY2RHbApkSmpRdG9SNk9JTWdRK1hCSURjZ3pnMlF1c2dLeTlxWUN0RGI2SnZYcVI0K3pQVldWU016UEdDaXlpU1ZENm1LCklzTnVUaHpCSGVQekVadTFaS1NpZnluM0JPUWFXaytkOHI2TlRLZTQxNVRON2pkM0NhWGY1ZTZJelUyaUkwYjcKUElzd2RUdjZkc1JoTjFkRU5vaFk4aGZ4ZzV4ckdzN3JzZnNEeitWcUdQaHYxekpzcGZsM3c5d1IwRW04QWRpZApDYW1ma01OeVZDdjhLSE5hTDBVd2xUTFovWGdoQ2Vkb0xHVW5LS0x2aHRrTTNmY1BHMW4rZTRJQjc1blF0b1VNCmRnbzlYQzJtbUFRSTIzUFpXdFI5MUpDYwotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    service:
      name: istiod-v110
      namespace: istio-system
      path: /inject
  failurePolicy: Fail
  name: rev.namespace.sidecar-injector.istio.io
  namespaceSelector:
    matchExpressions:
    - key: istio.io/rev
      operator: In
      values:
      - stable2
    - key: istio-injection
      operator: DoesNotExist
  objectSelector:
    matchExpressions:
    - key: sidecar.istio.io/inject
      operator: NotIn
      values:
      - "false"
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - CREATE
    resources:
    - pods
  sideEffects: None
- admissionReviewVersions:
  - v1beta1
  - v1
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMzakNDQWNhZ0F3SUJBZ0lSQU5GYmZObklMejhyNy9xdytidFNVazh3RFFZSktvWklodmNOQVFFTEJRQXcKR0RFV01CUUdBMVVFQ2hNTlkyeDFjM1JsY2k1c2IyTmhiREFlRncweU1EQTJNVGN3TVRVM01qRmFGdzB6TURBMgpNVFV3TVRVM01qRmFNQmd4RmpBVUJnTlZCQW9URFdOc2RYTjBaWEl1Ykc5allXd3dnZ0VpTUEwR0NTcUdTSWIzCkRRRUJBUVVBQTRJQkR3QXdnZ0VLQW9JQkFRQ3dmZnNKVkFDbnI1cjBSMHArcWdBRVFhWlVXMFR2M0dqTXkyYXAKelN2aEl3THFFSXNaWjE2WkxVZU1zaDEzYUFXTmRObDVkWkhZTmFZMHVJVzZsZlJ0bHJ6T2cwL295L3BJM2tBRgpwWEc2ajlXWkhmMVFVWkpHWnRFTFpXZDU1dmdwcTNtSHZ6SmhPZk0xT0RwMENnYkg0SVoxaHBsdnhPdnNFYkFxCjZ6Y21TMzNHc1ZkU2VKNjIwWGdYL2tBejlFRTFqVlVqUDV5eUJwYXlnQ3VVdHFCZ1JzbGhleWNkeEFvMTZhMVEKSTV3enZhTkdYa3IrQjhaaXFPUi9DWU5vdUhHbERSdk9JM25wY25JMnFkY2hKTkcwczA0S3ZJL0xNWWd4UytBbgpMdFp5VmJ2d1N3a1hkN2lxNjdPYlkrTkUxc3c2NGJZVEY2NUVwK0xtYnpITkgvVC9BZ01CQUFHakl6QWhNQTRHCkExVWREd0VCL3dRRUF3SUNCREFQQmdOVkhSTUJBZjhFQlRBREFRSC9NQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUIKQVFCT0NuRFV2RE4yRFEzWW9ILzdya3B1emF3SFYwUEE0aVhKUnhjSFZVTlpleXRtRUsyWGdVV2JHYURpY2RHbApkSmpRdG9SNk9JTWdRK1hCSURjZ3pnMlF1c2dLeTlxWUN0RGI2SnZYcVI0K3pQVldWU016UEdDaXlpU1ZENm1LCklzTnVUaHpCSGVQekVadTFaS1NpZnluM0JPUWFXaytkOHI2TlRLZTQxNVRON2pkM0NhWGY1ZTZJelUyaUkwYjcKUElzd2RUdjZkc1JoTjFkRU5vaFk4aGZ4ZzV4ckdzN3JzZnNEeitWcUdQaHYxekpzcGZsM3c5d1IwRW04QWRpZApDYW1ma01OeVZDdjhLSE5hTDBVd2xUTFovWGdoQ2Vkb0xHVW5LS0x2aHRrTTNmY1BHMW4rZTRJQjc1blF0b1VNCmRnbzlYQzJtbUFRSTIzUFpXdFI5MUpDYwotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    service:
      name: istiod-v110
      namespace: istio-system
      path: /inject
  failurePolicy: Fail
  name: rev.object.sidecar-injector.istio.io
  namespaceSelector:
    matchExpressions:
    - key: istio.io/rev
      operator: DoesNotExist
    - key: istio-injection
      operator: DoesNotExist
  objectSelector:
    matchExpressions:
    - key: sidecar.istio.io/inject
      operator: NotIn
      values:
      - "false"
    - key: istio.io/rev
      operator: In
      values:
      - stable2
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - CREATE
    resources:
    - pods
  sideEffects: None
```
