#!/bin/bash
docker stop arangodb
docker rm arangodb
rm -r arangodb
rm -r arangodb-apps
docker stop arangodb0
docker rm arangodb0
rm -r arangodb0
rm -r arangodb-apps0
docker stop arangodb1
docker rm arangodb1
rm -r arangodb1
rm -r arangodb-apps1
docker stop arangodb2
docker rm arangodb2
rm -r arangodb2
rm -r arangodb-apps2
