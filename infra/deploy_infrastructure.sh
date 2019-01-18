#!/bin/bash
BASEDIR=$(dirname $0)

echo "Deploying Kafka"
oc apply -f ${PWD}/${BASEDIR}/kafka/.

echo "Deploying ArangoDB"
oc apply -f ${PWD}/${BASEDIR}/arangodb/.

echo "Deploying InfluxDB"
oc apply -f ${PWD}/${BASEDIR}/influxdb/.

echo "Deploying Grafana"
oc apply -f ${PWD}/${BASEDIR}/grafana/.

echo "Deploying OpenBMPD"
sudo python ${PWD}/${BASEDIR}/openbmpd/deploy_openbmp.py

echo "Deploying Pipeline Ingress"
oc apply -f ${PWD}/${BASEDIR}/pipeline-ingress/.

echo "Deploying Telemetry"
python ${PWD}/${BASEDIR}/telemetry/deploy_telemetry.py

echo "Deploying Pipeline Egress"
oc apply -f ${PWD}/${BASEDIR}/pipeline-egress/.
