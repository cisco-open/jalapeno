#!/bin/bash
BASEDIR=$(dirname $0)

echo "Deploying API"
sh ${PWD}/${BASEDIR}/arangodb/deploy_foxx_service.sh
