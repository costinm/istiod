# Combining Helm and Kustomize

Helm flags:

```

--debug 
--dry-run

```

## Minimal chart

## Post-processing

```shell
helm upgrade --install example manifest/charts/example \
   --post-renderer ./kustomize 
```
```shell

helm template security-test ../example-chart \
   | ./kustomize  | kubectl apply -f -  
```
