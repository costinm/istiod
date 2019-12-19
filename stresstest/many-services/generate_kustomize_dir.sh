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
    sed -e "s/STRESSTESTREPLACEME/st${1:-}/g" -i $x;
  done

  kubectl kustomize "${gendir}" >> "${TMPDIR}/${1:-}.yaml"
  rm -rf "${gendir}"
}

seq -w 1 "${NUM_NAMESPACES:-5}" | while read index; do
  generate_template "${index}"
done
