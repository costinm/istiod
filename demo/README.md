### Istiod with TD -- Demo setup

This'll set up a test environment for Istiod + Traffic Director/NEG.

Due to some ordering requirements in the setup, you have to run the setup in three steps.

1) Cluster bootstrap -- this sets up all the CNIs, some empty mutating webhooks
   (which need to be patched with cert information), and other cluster-wide
   configuration

2) ASM setup -- this bootstraps the istiod instances (one standard, one with
   ASM)

3) Fortio (test client) -- this bootstraps three namespaces with sample
   applications. One runs non-ASM istiod and can demonstrate telemetry. The
   second runs istiod + asm. The third runs istiod + asm with SDS support.
   
To set up your cluster, run the following:

```
# CD to this directory if you haven't already
cd demo/

# Init the cluster
kubectl apply -k cluster-init/

# Init istiod/asm
kubectl apply -k asm-init/

# Init fortio
sleep 5 && kubectl apply -k fortio/
```

The delay ensures that the istiod instances set up in `asm-init` have time to
patch their mutatingwebhooks. If you don't wait, the fortio install step may fail.
