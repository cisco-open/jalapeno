#!/bin/sh
KUBE=$1
if [ -z "$1" ]
  then
    KUBE=kubectl
fi

echo "Shutting down Jalapeno"
${KUBE} delete -f ${PWD}/infra/service_account.yaml
${KUBE} delete namespace jalapeno
${KUBE} delete namespace jalapeno-collectors

echo "Deleting Persistent Volumes"
${KUBE} delete pv arangodb
${KUBE} delete pv arangodb-apps
${KUBE} delete pv pvkafka
${KUBE} delete pv pvzoo

echo "Deleting Data Stores"
sudo rm -rf /var/lib/kafka/data/topics
sudo rm -rf /var/lib/zookeeper/data
sudo rm -rf /var/lib/arangodb3-apps/_db
sudo rm -rf /var/lib/arangodb3/*
