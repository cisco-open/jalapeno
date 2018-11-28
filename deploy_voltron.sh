#!/bin/sh

### Deploying Voltron ###

### OpenShift login
oc login https://localhost:8443 -u admin -p admin -n voltron

### Creating the Voltron project
oc new-project voltron --description=voltron --display-name=voltron
sleep 3

### Deploying Infrastructure
sh infra/deploy_infrastructure.sh
sleep 20

### Deploying Services (vCollectors)
sh services/collectors/deploy_collectors.sh
sleep 20

### Deploying API
sh services/api/deploy_api.sh
