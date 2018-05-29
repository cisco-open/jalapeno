OC = /usr/local/bin/oc
MODULES = kafka arango framework topology collectors responders
VOLUMES = arangodb arangodb-apps pvkafka pvzoo
PROJECT = voltron

all: ${MODULES}

${MODULES}: ${PROJECT}
	${OC} apply -f $@/.

${PROJECT}:
	- ${OC} new-project ${PROJECT} --description=${PROJECT} --display-name=${PROJECT}

delete_project:
	${OC} delete project ${PROJECT}
	sleep 120

${VOLUMES}: delete_project
	${OC} delete pv $@

clean: ${VOLUMES}
	rm -rf /export/arangodb/{databases,journals,rocksdb,ENGINE,SERVER,SHUTDOWN}
	rm -rf /export/arangodb-apps/_db
	rm -rf /export/pvkafka/topics
	rm -rf /export/pvzoo/{data,log}

status:
	${OC} project ${PROJECT}
	${OC} get all

