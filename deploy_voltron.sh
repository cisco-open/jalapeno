#!/bin/sh
oc new-project voltron --description=voltron --display-name=voltron
sleep 3
sh infra/deploy_infrastructure.sh
sleep 15
sh services/deploy_services.sh
