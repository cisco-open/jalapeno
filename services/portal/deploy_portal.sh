#!/usr/bin/env bash
BASEDIR=$(dirname $0)
KUBE=$1
if [ -z "$1" ]
  then
    KUBE=kubectl
fi

${KUBE} create -f ${PWD}/${BASEDIR}/portal.yaml
${KUBE} create -f ${PWD}/${BASEDIR}/portal_svc_np.yaml
