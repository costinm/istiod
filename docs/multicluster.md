# Multicluster notes


## Official docs 

https://istio.io/latest/docs/setup/install/multicluster/multi-primary/

With helm and in safe mode - 

```yaml

 global:
      meshID: mesh1
      multiCluster:
        clusterName: cluster1
      network: network1
```
```shell

 istioctl x create-remote-secret \
 --context="${CTX_CLUSTER1}" \
 --name=cluster1 | \
 kubectl apply -f - --context="${CTX_CLUSTER2}"

```

## Best practice

- use 'canonical' cluster names - for example in GKE use the kubeconfig style.
- 
