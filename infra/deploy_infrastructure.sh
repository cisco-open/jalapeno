#!/bin/bash
BASEDIR=$(dirname $0)

KUBE=microk8s.kubectl

echo "Creating Jalapeno Namespace"
${KUBE} create -f ${PWD}/${BASEDIR}/namespace-jalapeno.json

echo "Setting up secret for docker.io"
${KUBE} create secret generic regcred --from-file=.dockerconfigjson=${HOME}/.docker/config.json --type=kubernetes.io/dockerconfigjson --namespace=jalapeno

echo "Deploying Kafka"
${KUBE} create -f ${PWD}/${BASEDIR}/kafka/.

echo "Deploying ArangoDB"
${KUBE} create -f ${PWD}/${BASEDIR}/arangodb/.

echo "Deploying InfluxDB"
${KUBE} create -f ${PWD}/${BASEDIR}/influxdb/.

echo "Deploying Grafana"
${KUBE} create -f ${PWD}/${BASEDIR}/grafana/.

echo "Deploying Pipeline Egress"
${KUBE} create -f ${PWD}/${BASEDIR}/pipeline-egress/.

echo "Finished deploying infra services"


