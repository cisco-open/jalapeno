#!/bin/bash
BASEDIR=$(dirname $0)
KUBE=$1
if [ -z "$1" ]
  then
    KUBE=kubectl
fi

echo "Deploying Topology AIO"
${KUBE} create -f ${PWD}/${BASEDIR}/topology-aio/topology-aio.yaml
