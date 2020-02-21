#!/bin/bash
BASEDIR=$(dirname $0)
echo "Starting deployment of telemetry services"
KUBE=microk8s.kubectl

echo "Creating jalapeno-telemetry Namespace"
${KUBE} create -f ${PWD}/${BASEDIR}/namespace-jalapeno-telemetry.json

echo "Deploying Pipeline Ingress"
${KUBE} create -f ${PWD}/${BASEDIR}/pipeline-ingress/.

echo "Deploying Openbmpd Collector"
${KUBE} create -f ${PWD}/${BASEDIR}/openbmpd/.

echo "Finished deploying telemetry services"
echo "Next configure routers!"
