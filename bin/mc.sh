
CTXTS=$(kubectl config get-contexts -o name |grep gke)

for i in ${CTXTS}; do
  IFS=_ read -ra p <<< $i
  echo $p
  echo $i
  istioctl x create-remote-secret --context $i --name ${p[1]}--${p[2]}--${p[3]} > $i.secret.yamls
done
