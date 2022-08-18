#!/bin/sh

### Deploying Minimal Jalapeno ###

### Deploying Minimal Infrastructure
sh infra/deploy_minimal_infrastructure.sh $1
sleep 5

### Deploying Minimal Collectors
#sh collectors/deploy_minimal_collectors.sh $1
#sleep 5

### Deploying Minimal Processors (Topology AIO)
sh processors/deploy_minimal_processors.sh $1

