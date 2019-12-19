
This script will create a kustomize folder that will generate a variable number
of instances of fortio (service, fortio server, and fortio client making
requests against it).

The number of instances created can be configured by the environment variable
"NUM_NAMESPACES". If not set, `generate_kustomize_dir.sh` will default to
creating 5 namespaces, each running a fortio stack.

The generated directory will be found in `./tmp/`. Each run of the script will
create a new directory with a different randomized suffix, to prevent
collisions. The directory used for each run will be echoed to STDOUT.

To apply the stress test to your cluster, run `kubectl apply -f
/path/to/directory`.

Warning: If you are generating a large number of namespaces, this may take some
time. Additionally, if you are creating a large number of namespaces on a small
cluster, you may get errors as the cluster tries to scale to your workload. It
is possible for a very large stress test kubectl apply to temporarily take down
the master nodes, as well -- this can be a very expensive process.

Use at your own risk!
