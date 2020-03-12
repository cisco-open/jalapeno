#!/bin/bash
BASEDIR=$(dirname $0)
KUBE=$1
if [ -z "$1" ]
  then
    KUBE=kubectl
fi

${KUBE} create -f ${PWD}/${BASEDIR}/topology/topology_dp.yaml
${KUBE} create -f ${PWD}/${BASEDIR}/l3vpn/l3vpn_dp.yaml
${KUBE} create -f ${PWD}/${BASEDIR}/epe-edges/epe_edges_collector_dp.yaml
${KUBE} create -f ${PWD}/${BASEDIR}/epe-paths/paths_collector_dp.yaml
${KUBE} create -f ${PWD}/${BASEDIR}/external-links-performance/external_links_performance_collector_dp.yaml
${KUBE} create -f ${PWD}/${BASEDIR}/internal-links-performance/internal_links_performance_collector_dp.yaml
