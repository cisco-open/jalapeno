#!/bin/bash

zip -r queries.zip queries
foxx server set voltron http://10.0.250.2:30852 -u root -P -D voltron
foxx install /queries --server voltron /home/centos/voltron/infra/arangodb/queries.zip
