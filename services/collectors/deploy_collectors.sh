#!/bin/bash
BASEDIR=$(dirname $0)
oc apply -f ${PWD}/${BASEDIR}/topology/.
oc apply -f ${PWD}/${BASEDIR}/epe-edges/.
oc apply -f ${PWD}/${BASEDIR}/epe-paths/.
oc apply -f ${PWD}/${BASEDIR}/external-links-performance/.
oc apply -f ${PWD}/${BASEDIR}/internal-links-performance/.
