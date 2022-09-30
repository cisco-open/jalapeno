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

ATTEMPTS=0
TIMEOUT=5
sleep 10
until $KUBE get all -n jalapeno | grep -o '[0-9]\/[0-9]' | grep -v '0\/1'
do
  ((ATTEMPTS=ATTEMPTS+1))
  if [ $ATTEMPTS -gt 12]; then
    echo "Cluster failed to come up"
    exit 1
  fi
  sleep $TIMEOUT
done
