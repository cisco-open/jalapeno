#!/bin/bash
BASEDIR=$(dirname $0)
oc apply -f ${PWD}/${BASEDIR}/bandwidth-responder/.
# oc apply -f ${PWD}/${BASEDIR}/latency-responder/.

