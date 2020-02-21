#!/usr/bin/env bash
BASEDIR=$(dirname $0)
KUBE=microk8s.kubectl

${KUBE} create -f ${PWD}/${BASEDIR}/portal.yaml
${KUBE} create -f ${PWD}/${BASEDIR}/portal_svc_np.yaml
