#!/bin/bash
BASEDIR=$(dirname $0)

echo "Creating Jalapeno Namespace"
kubectl create -f ${PWD}/${BASEDIR}/namespace-jalapeno.yaml

echo "Deploying Kafka"
kubectl create -f ${PWD}/${BASEDIR}/kafka/.

echo "Deploying ArangoDB"
kubectl create -f ${PWD}/${BASEDIR}/arangodb/.

echo "Deploying InfluxDB"
kubectl create -f ${PWD}/${BASEDIR}/influxdb/.

echo "Deploying Grafana"
kubectl create -f ${PWD}/${BASEDIR}/grafana/.

echo "Deploying Pipeline Egress"
kubectl create -f ${PWD}/${BASEDIR}/pipeline-egress/.

echo "Finished deploying infra services"


