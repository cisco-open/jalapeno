#!/bin/sh

### Deploying Jalapeno ###

### Deploying Infrastructure
sh infra/deploy_infrastructure.sh $1

### Deploying Collectors
sh collectors/deploy_collectors.sh $1

### Deploying Processors
sh processors/deploy_processors.sh $1

### Deploying API
sh api/deploy_api.sh $1
