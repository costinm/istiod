# Helper for telemetry bootstrap

Opinionated config for open telemetry and prometheus, for use in main().

- traces go to zpages or OTel collector if env is set.
- prometheus, expvar and otel metrics supported in libraries for metrics

## Accessing metrics and traces programmatically

For tests and debug tools it is useful to be able to access local telemetry.
Even in regular code it helps to know the current status.

