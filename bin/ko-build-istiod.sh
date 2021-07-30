#!/usr/bin/env bash

export KO_DOCKER_REPO=$(echo $IMAGE | cut -d: -f 1)
TAG=$(echo $IMAGE | cut -d: -f 2)

TARGET=$1

echo T $TARGET ARG $*
env

#export GOMAXPROCS=1

# .ko config is in istiod repo, to not polute istio/istio and
# to set custom base.
#export KO_CONFIG_PATH=$BUILD_CONTEXT/../istiod
ls ${KO_CONFIG_PATH}

export GOROOT=$(go env GOROOT)

cd ${BUILD_CONTEXT}

export GGCR_EXPERIMENT_ESTARGZ=1

echo ko publish --bare ${TARGET} -t $TAG
output=$(ko publish --bare ${TARGET} -t $TAG  | tee)

ref=$(echo $output | tail -n1)

