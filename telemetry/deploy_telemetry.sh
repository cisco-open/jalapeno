#!/bin/bash
BASEDIR=$(dirname $0)
echo "Starting deployment of telemetry services"

echo "Creating jalapeno-telemetry Namespace"
kubectl create -f ${PWD}/${BASEDIR}/namespace-jalapeno-telemetry.yaml

echo "Deploying Pipeline Ingress"
kubectl create -f ${PWD}/${BASEDIR}/pipeline-ingress/.

echo "Deploying Openbmpd Collector"
kubectl create -f ${PWD}/${BASEDIR}/openbmbd/.

echo "Finished deploying telemetry services"
echo "Next configure routers!"