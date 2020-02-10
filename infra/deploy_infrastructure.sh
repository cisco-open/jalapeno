#!/bin/bash
BASEDIR=$(dirname $0)

echo "Deploying Kafka"
kubectl -f ${PWD}/${BASEDIR}/kafka/.

echo "Deploying ArangoDB"
kubectl -f ${PWD}/${BASEDIR}/arangodb/.

echo "Deploying InfluxDB"
kubectl -f ${PWD}/${BASEDIR}/influxdb/.

echo "Deploying Grafana"
kubectl -f ${PWD}/${BASEDIR}/grafana/.

echo "Deploying OpenBMPD"
sudo python ${PWD}/${BASEDIR}/openbmpd/deploy_openbmp.py

echo "Deploying Pipeline Ingress"
oc apply -f ${PWD}/${BASEDIR}/pipeline-ingress/.

echo "Deploying Telemetry"
python3.6 ${PWD}/${BASEDIR}/telemetry/deploy_telemetry.py

echo "Deploying Pipeline Egress"
oc apply -f ${PWD}/${BASEDIR}/pipeline-egress/.
