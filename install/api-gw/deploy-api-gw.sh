#!/bin/bash

cd nginx-ingress-helm-operator
git checkout v1.0.0
make deploy IMG=nginx/nginx-ingress-operator:1.0.0
cp examples/*.yaml examples/deployment-oss-min/*.yaml ..
cd ..
$(KUBECTL) apply -f .
