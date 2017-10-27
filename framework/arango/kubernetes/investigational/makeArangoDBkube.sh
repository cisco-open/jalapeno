#!/bin/bash

# This scripts creates a yaml file to deploy an ArangoDB cluster to kubernetes.

NAME=arangodb
NRAGENTS=1
NRCOORDINATORS=1
NRDBSERVERS=2
#DOCKERIMAGE=arangodb/arangodb:latest
DOCKERIMAGE=registry.arangodb.com/arangodb/arangodb:3.1-devel
OUTPUT=arangodb_cluster.yaml
PROTO=tcp

function help() {
  echo "Usage: makeArangoDBkube.sh [options]"
  echo ""
  echo "Options:"
  echo "  -a/--agents NUMBER        (odd integer,  default: $NRAGENTS)"
  echo "  -c/--coordinators NUMBER  (integer >= 1, default: $NRCOORDINATORS)"
  echo "  -d/--dbservers NUMBER     (integer >= 2, default: $NRDBSERVERS)"
  echo "  -i/--image NAME           (name of Docker image, default: $DOCKERIMAGE)"
  echo "  -n/--name NAME            (name of ArangoDB cluster, default: $NAME)"
  echo "  -o/--output FILENAME      (name of output file, default: $OUTPUT)"
  echo "  -t/--tls BOOL             (if given, use TLS, default: false)"
}

while [[ ${1} ]] ; do
  case "${1}" in
    -a|--agents)
      NRAGENTS=${2}
      shift
      ;;
    -c|--coordinators)
      NRCOORDINATORS=${2}
      shift
      ;;
    -d|--dbservers)
      NRDBSERVERS=${2}
      shift
      ;;
    -i|--image)
      DOCKERIMAGE=${2}
      shift
      ;;
    -n|--name)
      NAME=${2}
      shift
      ;;
    -o|--output)
      OUTPUT=${2}
      shift
      ;;
    -t|--tls)
      if [ "${2}" = "true" ] ; then
        PROTO=ssl
      fi
      shift
      ;;
    -h|--help)
      help
      exit 1
      ;;
  esac

  if ! shift; then
    echo 'Missing parameter argument.' >&2
    exit 1
  fi
done

if [[ $(( $NRAGENTS % 2)) == 0 ]]; then
  echo "**ERROR: Number of agents must be odd! Bailing out."
  exit 1
fi

echo ======================
echo Planned configuration:
echo ======================
echo
echo "  Number of agents: $NRAGENTS"
echo "  Number of coordinators: $NRCOORDINATORS"
echo "  Number of dbservers: $NRDBSERVERS"
echo "  Docker image: $DOCKERIMAGE"
echo "  ArangoDB cluster name: $NAME"
echo "  Output file: $OUTPUT"
echo
echo Getting to work...

rm -f $OUTPUT
touch $OUTPUT

# Write out the deployments and services for the agents:
for i in $(seq 1 $NRAGENTS) ; do
  cat >>$OUTPUT <<EOF
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: ${NAME}-agent-${i}
spec:
  securityContext:
     fsGroup: 99
     runAsUser: 99
  replicas: 1
  template:
    metadata:
      labels:
        name: ${NAME}
        role: agent
        number: "${i}"
    spec:
      containers:
        - name: ${NAME}
          image: ${DOCKERIMAGE}
          ports:
            - containerPort: 8529
          volumeMounts:
            - mountPath: /var/lib/arangodb3
              name: arangodb
          args:
            - /usr/sbin/arangod
            - --server.authentication
            - "false"
            - --server.endpoint
            - ${PROTO}://0.0.0.0:8529
            - --agency.activate
            - "true"
            - --agency.size
            - "3"
            - --agency.supervision
            - "true"
            - --agency.my-address
            - ${PROTO}://${NAME}-agent-${i}.default.svc.cluster.local:8529
EOF
  if [[ $i > 1 ]] ; then
    im=$(expr $i - 1)
    for j in $(seq 1 $im) ; do
      cat >>$OUTPUT <<EOF
            - --agency.endpoint
            - ${PROTO}://${NAME}-agent-${j}.default.svc.cluster.local:8529
EOF
    done
  fi
  cat >>$OUTPUT <<EOF
      volumes:
        - name: arangodb
          emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: ${NAME}-agent-${i}
spec:
  ports:
    - port: 8529
      targetPort: 8529
  type: ClusterIP
  selector:
    name: ${NAME}
    role: agent
    number: "${i}"
EOF
done

# Write out the deployment for the dbservers:
cat >>$OUTPUT <<EOF
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: ${NAME}-dbs
spec:
  replicas: $NRDBSERVERS
  template:
    metadata:
      labels:
        name: ${NAME}
        role: dbserver
    spec:
      containers:
        - name: ${NAME}
          image: ${DOCKERIMAGE}
          env:
            - name: IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
          ports:
            - containerPort: 8529
          volumeMounts:
            - mountPath: /var/lib/arangodb3
              name: arangodb
          args:
            - /usr/sbin/arangod
            - --server.authentication
            - "false"
            - --server.endpoint
            - ${PROTO}://0.0.0.0:8529
            - --cluster.my-role
            - PRIMARY
            - --cluster.my-local-info
            - "\$(IP)"
            - --cluster.my-address
            - ${PROTO}://\$(IP):8529
EOF
for i in $(seq 1 $NRAGENTS) ; do
  cat >>$OUTPUT <<EOF
            - --cluster.agency-endpoint
            - ${PROTO}://${NAME}-agent-${i}.default.svc.cluster.local:8529
EOF
done
cat >>$OUTPUT <<EOF
      volumes:
        - name: arangodb
          emptyDir: {}
EOF

# Write out the deployment for the coordinators:
cat >>$OUTPUT <<EOF
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: ${NAME}-coords
spec:
  replicas: $NRCOORDINATORS
  template:
    metadata:
      labels:
        name: ${NAME}
        role: coordinator
    spec:
      containers:
        - name: ${NAME}
          image: ${DOCKERIMAGE}
          env:
            - name: IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
          ports:
            - containerPort: 8529
          volumeMounts:
            - mountPath: /var/lib/arangodb3
              name: arangodb
          args:
            - /usr/sbin/arangod
            - --server.authentication
            - "false"
            - --server.endpoint
            - ${PROTO}://0.0.0.0:8529
            - --cluster.my-role
            - COORDINATOR
            - --cluster.my-local-info
            - "\$(IP)"
            - --cluster.my-address
            - ${PROTO}://\$(IP):8529
EOF
for i in $(seq 1 $NRAGENTS) ; do
  cat >>$OUTPUT <<EOF
            - --cluster.agency-endpoint
            - ${PROTO}://${NAME}-agent-${i}.default.svc.cluster.local:8529
EOF
done
cat >>$OUTPUT <<EOF
      volumes:
        - name: arangodb
          emptyDir: {}
EOF

# Write out the service for the coordinators:
cat >>$OUTPUT <<EOF
---
apiVersion: v1
kind: Service
metadata:
  name: ${NAME}-coords
spec:
  ports:
    - port: 8529
      targetPort: 8529
  type: LoadBalancer
  selector:
    name: ${NAME}
    role: coordinator
EOF

echo Resulting YAML file has been written to $OUTPUT
echo Done.
