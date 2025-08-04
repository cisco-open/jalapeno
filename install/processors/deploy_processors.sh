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

echo "Deploying Linkstate Edge Processor"
${KUBE} create -f ${PWD}/${BASEDIR}/linkstate-edge/linkstate-edge.yaml
