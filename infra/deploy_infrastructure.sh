#!/bin/bash
BASEDIR=$(dirname $0)
oc apply -f ${PWD}/${BASEDIR}/kafka/.
oc apply -f ${PWD}/${BASEDIR}/arangodb/.
oc apply -f ${PWD}/${BASEDIR}/influxdb/.
oc apply -f ${PWD}/${BASEDIR}/grafana/.
oc apply -f ${PWD}/${BASEDIR}/openbmpd/.
#sleep 60
#python ${PWD}/${BASEDIR}/telemetry/deploy_telemetry.py
#oc apply -f ${PWD}/${BASEDIR}/pipeline/.
