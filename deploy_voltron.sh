#!/bin/sh

### Deploying Voltron ###

### OpenShift login
oc login https://localhost:8443 -u admin -p admin -n voltron

### Creating the Voltron project
oc new-project voltron --description=voltron --display-name=voltron
sleep 3

### Pulling images locally
docker pull ievoltron/topology:2.1
docker pull ievoltron/external-links-performance-collector:2.1.1
docker pull ievoltron/internal-links-performance-collector:2.1.1
docker pull ievoltron/api:0.0.1

### Deploying Infrastructure
sh infra/deploy_infrastructure.sh
sleep 20

### Deploying Services (vCollectors)
sh services/collectors/deploy_collectors.sh
sleep 20

### Deploying API
sh services/api/deploy_api.sh

### Deploying Portal
sh services/portal/deploy_portal.sh
