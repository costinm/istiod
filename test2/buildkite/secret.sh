#!/usr/bin/env bash

kubectl -n buildkite create secret generic buildkite-agent --from-literal token=${BUILDKITE_SECRET}
