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
### Assuming OpenBMP is configured on devices
### python ${PWD}/${BASEDIR}/openbmpd/configure_openbmp.py
### OpenBMP is running bare-metal, not containerized to avoid NATs etc.
### oc apply -f ${PWD}/${BASEDIR}/openbmpd/.
sudo python ${PWD}/${BASEDIR}/openbmpd/deploy_openbmp.py

echo "Deploying Telemetry"
python ${PWD}/${BASEDIR}/telemetry/deploy_telemetry.py

echo "Deploying Pipeline"
oc apply -f ${PWD}/${BASEDIR}/pipeline/.