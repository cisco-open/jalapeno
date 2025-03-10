#!/bin/bash
BASEDIR=$(dirname $0)

KUBE=$1
if [ -z "$1" ]
  then
    KUBE=kubectl
fi

echo "Deploying Jalapeno Collectors"

echo "Deploying Telegraf-Ingress Collector to collect network performance-metric data"
${KUBE} create -f ${PWD}/${BASEDIR}/telegraf-ingress/.

echo "Deploying GoBMP Collector to collect network topology data"
${KUBE} create -f ${PWD}/${BASEDIR}/gobmp/.

echo "Finished deploying Jalapeno Collectors"
