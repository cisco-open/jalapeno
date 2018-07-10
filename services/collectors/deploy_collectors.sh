#!/bin/bash
BASEDIR=$(dirname $0)
oc apply -f ${PWD}/${BASEDIR}/topology/.
oc apply -f ${PWD}/${BASEDIR}/path/.
oc apply -f ${PWD}/${BASEDIR}/bandwidth/.
# oc apply -f ${PWD}/${BASEDIR}/latency/.

