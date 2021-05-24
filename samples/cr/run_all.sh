# Deploy profiles and probers in all configured clusters

export TAG=${TAG:-asm}
export REV=${REV:-asm}
CTX=$(kubectl config get-contexts -o name | grep gke_ | xargs echo)

for i in ${CTX} ; do
  echo $i
  readarray -d_ -t gplc <<< "${i}_"
  # gke project location cluster
  PROJECT_ID=${gplc[1]} ZONE=${gplc[2]} CLUSTER=${gplc[3]} ENVEXTRA=ASM=1, make _run

  cat test/fortio.yaml | envsubst | kubectl --context $i apply -f -
done
