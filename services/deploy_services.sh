#!/bin/sh
BASEDIR=$(dirname $0)
sh ${PWD}/${BASEDIR}/collectors/deploy_collectors.sh
sh ${PWD}/${BASEDIR}/framework/deploy_framework.sh
#sh responders/deploy_responders.sh

