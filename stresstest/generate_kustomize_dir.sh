#!/usr/bin/env bash
set -euo pipefail 

mkdir -p "$(pwd)/tmp/"
TMPDIR="$(mktemp --directory --tmpdir="$(pwd)/tmp/" "generated_kustomize_XXX")"

echo "Created directory $TMPDIR to store generated kustomize manifests."

function generate_template {
  local gendir="${TMPDIR}/${1:-}"
  if [[ -d "${gendir}" ]]; then
     echo "ERROR: Directory ${gendir} already exists. Exiting!"
     exit 1
  fi
  cp -R "./template" "${gendir}"
  grep -i 'STRESSTESTREPLACEME' -r "${gendir}" -l | while read x; do
    sed -e "s/STRESSTESTREPLACEME/stressfortio${1:-}/g" -i $x;
  done
}

cat > "${TMPDIR}/kustomization.yaml" << EOF
bases:
EOF

seq 1 "${NUM_NAMESPACES:-5}" | while read index; do
  generate_template "${index}"
  echo "  - \"./${index}\"" >> "${TMPDIR}/kustomization.yaml"
done
