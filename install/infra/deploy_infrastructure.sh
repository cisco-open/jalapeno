#!/bin/bash
BASEDIR=$(dirname $0)

KUBE=$1
if [ -z "$1" ]
  then
    KUBE=kubectl
fi

echo "Creating Jalapeno Namespace"
${KUBE} create -f ${PWD}/${BASEDIR}/namespace-jalapeno.json

echo "Creating Jalapeno Service Account"
${KUBE} create -f ${PWD}/${BASEDIR}/service_account.yaml

echo "Setting up secret for docker.io"
${KUBE} create secret docker-registry regcred --docker-server="https://index.docker.io/v1/" --docker-username="jalapenoimageaccess" --docker-password="jalapeno2020" --docker-email="jalapeno-team@cisco.com" --namespace=jalapeno

echo "Deploying Kafka"
${KUBE} create -f ${PWD}/${BASEDIR}/kafka/.

echo "Deploying ArangoDB"
${KUBE} create -f ${PWD}/${BASEDIR}/arangodb/.

echo "Deploying InfluxDB"
${KUBE} create -f ${PWD}/${BASEDIR}/influxdb/.

echo "Deploying Grafana"
${KUBE} create -f ${PWD}/${BASEDIR}/grafana/.

echo "Deploying API"
${KUBE} create -f ${PWD}/${BASEDIR}/api/.

echo "Finished deploying infra services"


