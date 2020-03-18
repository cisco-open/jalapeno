#!/bin/sh

### Deploying Jalapeno ###

### Pulling images locally, update image version as necessary. Must be part of https://hub.docker.com/orgs/iejalapeno to pull images -- alternatively, build and upload images to personal image repos.
#docker pull iejalapeno/topology:2.1.2
#docker pull iejalapeno/external-links-performance-collector:2.1.1
#docker pull iejalapeno/internal-links-performance-collector:2.1.1
#docker pull iejalapeno/api:0.0.1.2
#docker pull iejalapeno/portal:0.0.1

### Deploying Infrastructure
sh infra/deploy_infrastructure.sh $1

### Deploying Collectors
sh collectors/deploy_collectors.sh $1

### Deploying Processors
sh processors/deploy_processors.sh $1

### Deploying API
sh api/deploy_api.sh $1
