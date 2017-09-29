#!/bin/bash
mkdir arangodb
mkdir arangodb-apps
docker run \
    --name arangodb \
    -e ARANGO_ROOT_PASSWORD=voltron \
    -p 8529:8529 \
    -v `pwd`/arangodb:/var/lib/arangodb \
    -v `pwd`/arangodb-apps:/var/lib/arangodb-apps\
    -d \
    arangodb/arangodb:3.1.26
