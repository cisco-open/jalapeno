#!/bin/bash
BASEDIR=$(dirname $0)

printf "Deploying Kafka"
oc apply -f ${PWD}/${BASEDIR}/kafka/.

printf "Deploying ArangoDB"
oc apply -f ${PWD}/${BASEDIR}/arangodb/.

printf "Deploying InfluxDB"
oc apply -f ${PWD}/${BASEDIR}/influxdb/.

printf "Deploying Grafana"
oc apply -f ${PWD}/${BASEDIR}/grafana/.

printf "Deploying OpenBMPD"
python ${PWD}/${BASEDIR}/openbmpd/configure_openbmp.py
oc apply -f ${PWD}/${BASEDIR}/openbmpd/.

printf "Deploying Telemetry"
python ${PWD}/${BASEDIR}/telemetry/deploy_telemetry.py

printf "Deploying Pipeline"
oc apply -f ${PWD}/${BASEDIR}/pipeline/.
