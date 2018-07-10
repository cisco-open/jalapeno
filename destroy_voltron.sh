#!/bin/sh -x

### Destroying Voltron ###

oc delete project voltron
sleep 30
oc delete pv arangodb 
oc delete pv arangodb-apps 
oc delete pv pvkafka 
oc delete pv pvzoo

sudo rm -rf /export/arangodb/{databases,journals,rocksdb,ENGINE,SERVER,SHUTDOWN}
sudo rm -rf /export/arangodb-apps/_db
sudo rm -rf /export/pvkafka/topics
sudo rm -rf /export/pvzoo/{data,log}

sh infra/destroy_infrastructure.sh
