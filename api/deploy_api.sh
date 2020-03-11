#!/usr/bin/env bash
BASEDIR=$(dirname $0)
KUBE=$1
if [ -z "$1" ]
  then
    KUBE=kubectl
fi

${KUBE} create -f ${PWD}/${BASEDIR}/api.yaml
${KUBE} create -f ${PWD}/${BASEDIR}/api_svc_np.yaml
