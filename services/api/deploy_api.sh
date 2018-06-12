#!/bin/bash
BASEDIR=$(dirname $0)

echo "Deploying APIs"
sh ${PWD}/${BASEDIR}/arangodb/deploy_foxx_service.sh

