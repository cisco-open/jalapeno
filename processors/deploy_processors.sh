#!/bin/bash
BASEDIR=$(dirname $0)
KUBE=$1
if [ -z "$1" ]
  then
    KUBE=kubectl
fi

${KUBE} create -f ${PWD}/${BASEDIR}/topology2/deployment/topology_dp.yaml

#${KUBE} create -f ${PWD}/${BASEDIR}/topology/deployment/topology_dp.yaml
#${KUBE} create -f ${PWD}/${BASEDIR}/lsv4/lsv4_dp.yaml
#${KUBE} create -f ${PWD}/${BASEDIR}/lsv6/lsv6_dp.yaml
#${KUBE} create -f ${PWD}/${BASEDIR}/l3vpn/l3vpn_dp.yaml
#${KUBE} create -f ${PWD}/${BASEDIR}/lsv4-performance/lsv4_performance_dp.yaml
#${KUBE} create -f ${PWD}/${BASEDIR}/lsv6-performance/lsv6_performance_dp.yaml

