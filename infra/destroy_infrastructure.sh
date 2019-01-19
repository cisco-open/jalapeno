#!/bin/bash
BASEDIR=$(dirname $0)

echo "Stopping OpenBMP"
sudo docker kill openbmp_collector
sudo docker rm openbmp_collector

echo "Stopping Telemetry"
python3.6 ${PWD}/${BASEDIR}/telemetry/remove_telemetry.py

