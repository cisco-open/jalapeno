# Jalapeno Installation
Jalapeno has been primarily developed, tested, and operated on Ubuntu 18.04 and 20.04 Kubernetes environments (bare-metal, VM, or cloud). Recommended VM sizing for a test lab is 4 vCPU, 16GB memory, and 50G of disk.  If deploying in production or a test environment with large table sizes (full Internet table, 250k + internal or vpn prefixes), then we recommend a bare metal K8s cluster with two or more nodes. 

Note: the Jalapeno installation script by default will pull a telemetry stack consisting of Telegraf, Influx, and Kafka images (the TIK stack).  If you would like to integrate Jalapeno's BMP/Topology/GraphDB elements with an existing telemetry stack simply comment out the TIK stack elements in the shell script.

Instructions for installing Kubernetes: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/

Users who do not have a full Kubernetes or GKE deployment can get up and running quite quickly with Microk8s [Installing K8s](docs/K8s_installation.md). (_Note that the DNS service must be enabled before deploying Jalapeno: `microk8s enable dns`._)

### Installing Jalapeno

1. Clone this repo and `cd` into the folder: `git clone <repo> && cd jalapeno`

2. Use the `deploy_jalapeno.sh` script. This will start the collectors, the Jalapeno infra images, and the topology and linkstate-edge processors.

   ```bash
   ./deploy_jalapeno.sh [path_to_kubectl]

   ```
  (_Note: if you're using a nonstandard kubectl, you need to pass the appropriate command to this script. For example, with microk8s:_ `./deploy_jalapeno.sh microk8s.kubectl`)

3. Check that all containers are up using: `kubectl get all --all-namespaces` or on a per-namespace basis:
```
kubectl get all -n jalapeno
kubectl get all -n jalapeno-collectors
```
Output
```
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
Collectors
```
NAME                                               READY   STATUS    RESTARTS   AGE
pod/gobmp-f8bf8d6d5-nvwn8                          1/1     Running   2          3m42s
pod/telegraf-ingress-deployment-56867cf9b4-62snv   1/1     Running   1          3m46s

NAME                          TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)                          AGE
service/gobmp                 NodePort   10.97.107.27    <none>        5000:30511/TCP,56767:30767/TCP   3m43s
service/telegraf-ingress-np   NodePort   10.97.218.162   <none>        57400:32400/TCP                  3m46s

NAME                                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/gobmp                         1/1     1            1           3m42s
deployment.apps/telegraf-ingress-deployment   1/1     1            1           3m46s

NAME                                                     DESIRED   CURRENT   READY   AGE
replicaset.apps/gobmp-f8bf8d6d5                          1         1         1       3m42s
replicaset.apps/telegraf-ingress-deployment-56867cf9b4   1         1         1       3m46s

```

4. Configure routers in the network to stream telemetry and BMP data to the Jalapeno cluster. Jalapeno's default MDT port is 32400 and the BMP port is 30511.  Generally we would setup MDT on all routers and BMP only on route reflectors and any routers with external peering sessions.

   1. Example destination group for MDT: **Note: you may need to set TPA mgmt**

      ```shell
       destination-group jalapeno
        address-family ipv4 <server-ip> port 32400
         encoding self-describing-gpb
         protocol grpc no-tls
        !
       !
      ```

   2. Example of IOS-XR BMP config:

      ```shell
      bmp server 1
       host <server-ip> port 30511
       description jalapeno GoBMP
       update-source MgmtEth0/RP0/CPU0/0
       flapping-delay 60
       initial-delay 5
       stats-reporting-period 60
       initial-refresh delay 30 spread 2
      !
      router bgp 65000
       neighbor 172.31.101.4
       bmp-activate server 1
      ```

## Destroying Jalapeno

Jalapeno can also be destroyed using the script.

1. Use the `destroy_jalapeno.sh` script. Will remove both namespaces jalapeno and jalapeno-collectors and all associated services/pods/deployments/etc. and it will remove all the persistent volumes associated with kafka and arangodb.

   ```shell
   destroy_jalapeno.sh kubectl
   ```


