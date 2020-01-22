#!/bin/sh
echo "Logging into OpenShift"
oc login https://localhost:8443 -u admin -p admin -n jalapeno

echo "Shutting down Jalapeno"
oc delete project jalapeno

echo "Please wait"
sleep 60

echo "Deleting Persistent Volumes"
oc delete pv arangodb
oc delete pv arangodb-apps
oc delete pv pvkafka
oc delete pv pvzoo

echo "Removing data logs"
sudo rm -rf /export/arangodb/*
sudo rm -rf /export/arangodb-apps/_db
sudo rm -rf /export/pvkafka/topics
sudo rm -rf /export/pvzoo/{data,log}

echo "Deleting OpenBMP and stopping Telemetry"
sh infra/destroy_infrastructure.sh
