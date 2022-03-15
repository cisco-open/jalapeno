#!/bin/sh

### Deploying Jalapeno ###

### Deploying Infrastructure
sh infra/deploy_infrastructure.sh $1
sleep 5

### Deploying Collectors
sh collectors/deploy_collectors.sh $1
sleep 5

### Deploying Processors
sh processors/deploy_processors.sh $1

