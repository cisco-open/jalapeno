#!/usr/bin/env bash
BASEDIR=$(dirname $0)
oc apply -f ${PWD}/${BASEDIR}/api.yml
oc apply -f ${PWD}/${BASEDIR}/api_svc_np.yaml
