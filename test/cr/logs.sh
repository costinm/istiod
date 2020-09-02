#/bin/bash

export REV=${REV:-asm}
export CLUSTER=${CLUSTER:-big1}
export SVC=${SVC:-istiod-costin-asm1-${CLUSTER}-${REV}}

gcloud logging read --project wlhe-cr \
  'resource.type = "cloud_run_revision" AND resource.labels.service_name = "'${SVC}'"' --limit=1000 --format json | jq '.[] | .labels.instanceId[-10:] + " " + (if .textPayload then .textPayload else (.httpRequest | tostring) end)' -r | tac
