#!/usr/bin/env bash
BASEDIR=$(dirname $0)
KUBE=microk8s.kubectl

${KUBE} create -f ${PWD}/${BASEDIR}/api.yaml
${KUBE} create -f ${PWD}/${BASEDIR}/api_svc_np.yaml
