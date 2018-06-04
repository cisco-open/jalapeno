#!/bin/sh
#oc login https://localhost:8443 -u admin -p admin -n voltron
#oc new-project voltron --description=voltron --display-name=voltron
#sleep 3
#sh infra/deploy_infrastructure.sh
#sleep 15
#sh services/deploy_services.sh
sh services/collectors/deploy_collectors.sh
sh services/framework/deploy_framework.sh
#sh services/responders/deploy_responders.sg
