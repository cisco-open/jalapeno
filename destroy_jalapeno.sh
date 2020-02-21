#!/bin/sh
KUBE=microk8s.kubectl

echo "Shutting down Jalapeno"
${KUBE} delete namespace jalapeno
${KUBE} delete namespace jalapeno-namespace

echo "Please wait"
sleep 60

echo "Deleting Persistent Volumes"
${KUBE} delete pv arangodb
${KUBE} delete pv arangodb-apps
${KUBE} delete pv pvkafka
${KUBE} delete pv pvzoo
