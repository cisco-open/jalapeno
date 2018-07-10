#!/bin/bash
mkdir arangodb
mkdir arangodb3-apps
docker run \
    --name arangodb \
    -e ARANGO_ROOT_PASSWORD=voltron \
    -p 8529:8529 \
    -d \
    arangodb/arangodb:3.2.2
