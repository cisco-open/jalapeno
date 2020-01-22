#!/bin/sh

### Deploying Jalapeno ###

### OpenShift login
oc login https://localhost:8443 -u admin -p admin -n jalapeno

### Creating the Jalapeno project
oc new-project jalapeno --description=jalapeno --display-name=jalapeno
sleep 3

### Pulling images locally, update image version as necessary. Must be part of https://hub.docker.com/orgs/iejalapeno to pull images -- alternatively, build and upload images to personal image repos.
docker pull iejalapeno/topology:2.1.2
docker pull iejalapeno/external-links-performance-collector:2.1.1
docker pull iejalapeno/internal-links-performance-collector:2.1.1
docker pull iejalapeno/api:0.0.1.2
docker pull iejalapeno/portal:0.0.1

### Deploying Infrastructure
sh infra/deploy_infrastructure.sh
sleep 20

### Deploying Services (vCollectors)
sh services/collectors/deploy_collectors.sh
sleep 20

### Deploying API
sh services/api/deploy_api.sh

### Deploying Portal
#sh services/portal/deploy_portal.sh
