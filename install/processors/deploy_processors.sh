#!/bin/bash
BASEDIR=$(dirname $0)
KUBE=$1
if [ -z "$1" ]
  then
    KUBE=kubectl
fi

echo "Deploying Topology"
${KUBE} create -f ${PWD}/${BASEDIR}/topology/topology.yaml

echo "Deploying Telegraf-Egress"
${KUBE} create -f ${PWD}/${BASEDIR}/telegraf-egress/.

echo "Deploying LS Link-Node Edge Processor"
${KUBE} create -f ${PWD}/${BASEDIR}/lslinknode-edge/lslinknode-edge.yaml

echo "Deploying IGP Graph Processor"
${KUBE} create -f ${PWD}/${BASEDIR}/graph-processors/igp-graph.yaml

echo "Deploying IPv4 Graph Processor"
${KUBE} create -f ${PWD}/${BASEDIR}/graph-processors/ipv4-graph.yaml

echo "Deploying IPv6 Graph Processor"
${KUBE} create -f ${PWD}/${BASEDIR}/graph-processors/ipv6-graph.yaml
