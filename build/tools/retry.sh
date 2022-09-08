#!/bin/bash

if [ -z "$1" ]; then
	echo "No command provided"
	exit 1
else
	command="$*"
fi
x=0
timeout=5
while true; do
	$command
	if [ $? -eq 0 ]; then
		exit 0
	fi
	((x=x+1))
	if [ $x -gt 12 ]; then
		echo "Command failed: $command"
		exit 1
	fi
	sleep $timeout
done
