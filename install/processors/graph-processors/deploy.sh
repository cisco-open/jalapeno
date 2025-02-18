#!/bin/sh

kubectl apply -f igp-graph.yaml
echo "Waiting for 10 seconds for igp-graph to be ready"
sleep 10
kubectl apply -f ipv4-graph.yaml
kubectl apply -f ipv6-graph.yaml
kubectl apply -f srv6-localsids.yaml
kubectl apply -f api-deployment.yaml
kubectl apply -f api-svc.yaml
kubectl apply -f ui-deployment.yaml
kubectl apply -f ui-svc.yaml

