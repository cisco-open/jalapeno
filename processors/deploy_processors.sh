#!/bin/bash
BASEDIR=$(dirname $0)
KUBE=$1
if [ -z "$1" ]
  then
    KUBE=kubectl
fi

${KUBE} create -f ${PWD}/${BASEDIR}/topology/topology_dp.yaml
${KUBE} create -f ${PWD}/${BASEDIR}/ls/ls_dp.yaml
${KUBE} create -f ${PWD}/${BASEDIR}/l3vpn/l3vpn_dp.yaml
${KUBE} create -f ${PWD}/${BASEDIR}/ls-performance/ls_performance_dp.yaml
