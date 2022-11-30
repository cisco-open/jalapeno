#!/bin/bash

KUBE=$1
if [ -z "$1" ]
  then
    KUBE=kubectl
fi

TEST=`which $KUBE`
if [ "$?" -eq 1 ]; then
    echo "$KUBE not found, exiting..."
    exit 1
fi

${KUBE} delete namespace jalapeno-api-gw nginx-ingress
cd nginx-ingress-helm-operator
make -f Makefile.jalapeno undeploy
make -f Makefile.jalapeno uninstall
#${KUBE} delete namespace nginx-ingress-operator-system
