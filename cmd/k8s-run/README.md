# Run an app in a k8s-like environment

This is a simple command that:
- connects to K8S, based on env variables (currently including GKE-specific code, using a GCP or CloudRun GSA)
- creates an environment similar with a istio-proxy pod
- execute the reminder of the command
- periodically refreshes the K8S token and other resources

Use case is running istio-agent or other applications expecting a K8S pod environment - without having to use
a shell script or modify the application. When running in K8S - the enviroment is setup and maintained by kubelet.
This app provides a minimal replacement.

## Configuration

Environment variables are used to locate the K8S cluster:

- CLUSTER - name of the cluster.

## Environment



## Dev environment

- when running as non-root: 
    - iptables will not be set
    - envoy will run with the current UID
    - all files will be created relative to current dir.

If the command is not running as root, the environment will be created based on the current dir. Istio agent 
typically runs in "/", and uses './path' to locate the files. 
