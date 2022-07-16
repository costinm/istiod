# App startup and configuration

The common methods for configuration are:

1. Command line flags. This is best for CLI tools, but not ideal for long lived servers. 
   - Can be used with auto-completion
   - Require restart
   - Errors if unknown flags are used - good for CLI, really bad for server upgrades, ties the config to the version
   - hard to patch in k8s/docker - ordering, part of the 'command' 

2. Environment variables
    - similar with CLI, require restart
    - less repetition - can be set once, no need to pass it on each command (example: KUBE_CONFIG)
    - better for servers, more robust on upgrade.
    - easy to patch. ConfigMaps can be used to mount additional envs.

3. Config files (including ConfigMaps)
    - Support dynamic update (via k8s watcher or file watcher)
    - if 'flat', easy to patch

4. CRDs
    - Require using the API server library, can't be mounted
    - Rest is similar with ConfigMaps
    - Can have static validation, autocomplete, etc
