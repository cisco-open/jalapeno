#!/bin/sh
oc login https://localhost:8443 -u admin -p admin -n voltron
oc new-project voltron --description=voltron --display-name=voltron
sleep 3
sh infra/deploy_infrastructure.sh
sleep 30
sh services/collectors/deploy_collectors.sh
sh services/framework/deploy_framework.sh
sleep 30
sh infra/deploy_api.sh

#sh services/responders/deploy_responders.sh
