#!/bin/sh

### Deploying Minimal Jalapeno ###

### Deploying Minimal Infrastructure
sh infra/deploy_minimal_infrastructure.sh $1
sleep 5

### Deploying Minimal Collectors
#sh collectors/deploy_minimal_collectors.sh $1
#sleep 5

### Deploying Processors
#sh processors/deploy_processors.sh $1

