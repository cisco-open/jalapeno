#!/bin/bash
BASEDIR=$(dirname $0)
echo "Starting deployment of telemetry services"
KUBE=$1
if [ -z "$1" ]
  then
    KUBE=kubectl
fi

echo "Creating jalapeno-telemetry Namespace"
${KUBE} create -f ${PWD}/${BASEDIR}/namespace-jalapeno-collectors.json

echo "Deploying Pipeline Ingress"
${KUBE} create -f ${PWD}/${BASEDIR}/pipeline-ingress/.

echo "Deploying Openbmpd Collector"
${KUBE} create -f ${PWD}/${BASEDIR}/openbmpd/.

echo "Finished deploying telemetry services"
echo "Next configure routers!"
