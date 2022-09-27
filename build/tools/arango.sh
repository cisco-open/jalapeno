#!/bin/bash
COLLECTION="unicast_prefix_v4"
TOKEN=$(curl -s -X POST -H 'Accept: application/json' -H 'Content-Type: application/json' --data '{"username":"root","password":"jalapeno"}' http://127.0.0.1:30852/_open/auth | jq -r '.jwt')
#echo $TOKEN

check_prefixes () {
  PREFIXES=$(curl -H 'Accept: application/json' -H "Authorization: Bearer ${TOKEN}" http://127.0.0.1:30852/_db/jalapeno/_api/collection/${COLLECTION}/count | jq -r '.count')
  [ $PREFIXES -eq $1 ] && return
}


check_prefixes 6

TEST=`which curl`
if [ "$?" -eq 1 ]; then
    echo "curl not found, exiting..."
    exit 1
fi

ATTEMPTS=0
TIMEOUT=20
until check_prefixes 6
do
  ((ATTEMPTS=ATTEMPTS+1))
  if [ $ATTEMPTS -gt 12 ]; then
    echo "Failed to find routes in Arango after 4 mins"
    exit 1
  fi
  sleep $TIMEOUT
done
