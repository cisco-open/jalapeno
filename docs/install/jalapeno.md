# Jalapeno Installation

!!! warning
    A Kubernetes installation is required to continue. If you don't have a running environment, please follow the steps [here](kubernetes.md) to get set up.

### Installing Jalapeno

???+ tip
    The Jalapeno installation script by default will pull a telemetry stack consisting of Telegraf, Influx, and Kafka images (the TIK stack). If you would like to  integrate Jalapeno's BMP/Topology/GraphDB elements with an existing telemetry stack simply comment out the TIK stack elements in the shell script.

1. Clone the Jalapeno repo and `cd` into the folder:

    ```bash
    git clone https://github.com/cisco-open/jalapeno.git && cd jalapeno/install
    ```

2. Use the `deploy_jalapeno.sh` script. This will start the collectors, the Jalapeno infra images, and the topology and linkstate-edge processors.

    ```bash
    ./deploy_jalapeno.sh [path_to_kubectl]
 
    ```

    !!! tip
        If you're using a nonstandard kubectl, you need to pass the appropriate command to this script.

        For example, with microk8s: `./deploy_jalapeno.sh microk8s.kubectl`

### Validation

Validate that all containers are started & running. Using: `kubectl get all --all-namespaces` or on a per-namespace basis:

```bash
kubectl get all -n jalapeno
```

Expected Output for `jalapeno` Namespace:

```{ .text .no-copy }
NAME                                              READY   STATUS    RESTARTS   AGE
pod/arangodb-0                                    1/1     Running   0          9d
pod/grafana-deployment-579c5f75bb-7g7bk           1/1     Running   0          9d
pod/influxdb-0                                    1/1     Running   0          9d
pod/kafka-0                                       1/1     Running   0          9d
pod/linkstate-edge-66fb9b8fb7-skmsc               1/1     Running   0          6d22h
pod/telegraf-egress-deployment-55cbff896c-vf26q   1/1     Running   3          9d
pod/topology-6fdd6ccc8b-6mcpc                     1/1     Running   0          8d
pod/zookeeper-0                                   1/1     Running   0          9d

NAME                  TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
service/arango-np     NodePort    10.152.183.143   <none>        8529:30852/TCP               9d
service/arangodb      ClusterIP   10.152.183.142   <none>        8529/TCP                     9d
service/broker        ClusterIP   10.152.183.247   <none>        9092/TCP                     9d
service/grafana       ClusterIP   10.152.183.62    <none>        3000/TCP                     9d
service/grafana-np    NodePort    10.152.183.124   <none>        3000:30300/TCP               9d
service/influxdb      ClusterIP   10.152.183.197   <none>        8086/TCP                     9d
service/influxdb-np   NodePort    10.152.183.68    <none>        8086:30308/TCP               9d
service/kafka         NodePort    10.152.183.160   <none>        9094:30092/TCP               9d
service/zookeeper     ClusterIP   10.152.183.36    <none>        2888/TCP,3888/TCP,2181/TCP   9d

NAME                                         READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/grafana-deployment           1/1     1            1           9d
deployment.apps/linkstate-edge               1/1     1            1           6d22h
deployment.apps/telegraf-egress-deployment   1/1     1            1           9d
deployment.apps/topology                     1/1     1            1           8d

NAME                                                    DESIRED   CURRENT   READY   AGE
replicaset.apps/grafana-deployment-579c5f75bb           1         1         1       9d
replicaset.apps/linkstate-edge-66fb9b8fb7               1         1         1       6d22h
replicaset.apps/telegraf-egress-deployment-55cbff896c   1         1         1       9d
replicaset.apps/topology-6fdd6ccc8b                     1         1         1       8d

NAME                         READY   AGE
statefulset.apps/arangodb    1/1     9d
statefulset.apps/influxdb    1/1     9d
statefulset.apps/kafka       1/1     9d
statefulset.apps/zookeeper   1/1     9d
```

### Device Config

Configure routers in the network to stream telemetry and BMP data to the Jalapeno cluster.

Instructions can be found under the [Device Config](../device-config/index.md) section.

## Destroying Jalapeno

Jalapeno can also be destroyed using the script.

1. Use the `destroy_jalapeno.sh` script. This will remove the `jalapeno` namespace and all associated services/pods/deployments/etc. This will also remove all the persistent volumes associated with Kafka and Arangodb.

   ```bash
   ./destroy_jalapeno.sh kubectl
   ```
