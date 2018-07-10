#!/bin/sh

### Deploying Voltron ###

### OpenShift login
oc login https://localhost:8443 -u admin -p admin -n voltron

### Creating the Voltron project
oc new-project voltron --description=voltron --display-name=voltron
sleep 3

### Deploying Infrastructure
sh infra/deploy_infrastructure.sh
sleep 15

### Deploying Services (vCollectors)
sh services/collectors/deploy_collectors.sh
sh services/framework/deploy_framework.sh
sleep 15

### Deploying Services (vResponders)
sh services/responders/deploy_responders.sh

### Deploying API
sh services/api/deploy_api.sh
