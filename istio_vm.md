# Notes on Istio for VM, Mesh Expansion

## Build and test

```shell script


BUILD_WITH_CONTAINER=1 make deb

# Creates a docker image where it installs the .deb, using docker directly
BUILD_WITH_CONTAINER=0 make deb/docker

# Interactive shell into the 'VM'
BUILD_WITH_CONTAINER=0 make deb/run/docker

```


