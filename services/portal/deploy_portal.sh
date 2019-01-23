#!/usr/bin/env bash
BASEDIR=$(dirname $0)
oc apply -f ${PWD}/${BASEDIR}/portal.yaml
oc apply -f ${PWD}/${BASEDIR}/portal_svc_np.yaml
