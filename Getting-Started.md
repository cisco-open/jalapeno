# Jalapeno Installation Guide
Jalapeno has been primarily developed, tested, and operated on Ubuntu 18.04 (bare-metal or VM), and Google Kubernetes Engine. Minimum VM sizing for a test lab is 2 vCPU, 4GB memory, and 8G of disk.  If deploying in an environment with large table sizes (full Internet table, 100k + internal or vpn prefixes), then we recommend bare metal or a VM with at least 4 vCPU, 16GB memory, and 40G of disk.

Users who do not have a full Kubernetes or GKE deployment can get up and running quite quickly with Microk8s [Installing MicroK8s](docs/MicroK8s_installation.md)

### Installing Jalapeno

1. Clone this repo and `cd` into the folder: `git clone <repo> && cd jalapeno`

2. Use the `deploy_jalapeno.sh` script. This will start the collectors and all jalapeno infra and the Topology processor on the single node.

Note: if using Microk8s you may need to put a 'microk8s.kubectl' the commands referenced below

   ```bash
   ./deploy_jalapeno.sh
   or
   ./deploy_jalapeno.sh microk8s.kubectl
   ```

3. Check that all containers are up using: `kubectl get all --all-namespaces` or on a per-namespace basis:
```
kubectl get all -n jalapeno
kubectl get all -n jalapeno-collectors
```
Output
```
kubectl get all -n jalapeno 

NAME                                              READY   STATUS        RESTARTS   AGE
pod/arangodb-0                                    1/1     Running       0          9d
pod/demo-l3vpn-processor-675b658cb7-np2kk         1/1     Terminating   3          31m
pod/demo-lsv4-perf-processor-78c5bfc5fb-vch7r     1/1     Terminating   3          31m
pod/demo-lsv4-processor-78dc64c9f5-xjn2q          1/1     Terminating   5          31m
pod/demo-lsv6-processor-7df649ff6f-gj2j6          1/1     Terminating   2          31m
pod/grafana-deployment-579c5f75bb-j6zwr           1/1     Running       0          9d
pod/influxdb-0                                    1/1     Running       0          9d
pod/kafka-0                                       1/1     Running       0          9d
pod/telegraf-egress-deployment-55cbff896c-7mqlh   1/1     Running       5          9d
pod/topology-6db7dc4fc4-rj5p5                     1/1     Running       0          33m
pod/zookeeper-0                                   1/1     Running       1          9d

NAME                  TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
service/arango-np     NodePort    10.152.183.34    <none>        8529:30852/TCP               9d
service/arangodb      ClusterIP   10.152.183.103   <none>        8529/TCP                     9d
service/broker        ClusterIP   10.152.183.11    <none>        9092/TCP                     9d
service/grafana       ClusterIP   10.152.183.186   <none>        3000/TCP                     9d
service/grafana-np    NodePort    10.152.183.83    <none>        3000:30300/TCP               9d
service/influxdb      ClusterIP   10.152.183.9     <none>        8086/TCP                     9d
service/influxdb-np   NodePort    10.152.183.176   <none>        8086:30308/TCP               9d
service/kafka         NodePort    10.152.183.109   <none>        9092:30092/TCP               9d
service/zookeeper     ClusterIP   10.152.183.93    <none>        2888/TCP,3888/TCP,2181/TCP   9d

NAME                                         READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/grafana-deployment           1/1     1            1           9d
deployment.apps/telegraf-egress-deployment   1/1     1            1           9d
deployment.apps/topology                     1/1     1            1           33m

NAME                                                    DESIRED   CURRENT   READY   AGE
replicaset.apps/grafana-deployment-579c5f75bb           1         1         1       9d
replicaset.apps/telegraf-egress-deployment-55cbff896c   1         1         1       9d
replicaset.apps/topology-6db7dc4fc4                     1         1         1       33m

NAME                         READY   AGE
statefulset.apps/arangodb    1/1     9d
statefulset.apps/influxdb    1/1     9d
statefulset.apps/kafka       1/1     9d
statefulset.apps/zookeeper   1/1     9d
```
Collectors
```
kubectl get all -n jalapeno-collectors
NAME                                              READY   STATUS    RESTARTS   AGE
pod/openbmpd-0                                    1/1     Running   1          40h
pod/telegraf-ingress-deployment-ddfc8ff47-66n9j   1/1     Running   2          40h

NAME                          TYPE       CLUSTER-IP       EXTERNAL-IP   PORT(S)           AGE
service/openbmpd-np           NodePort   10.152.183.23    <none>        5000:30555/TCP    40h
service/telegraf-ingress-np   NodePort   10.152.183.168   <none>        57400:32400/TCP   40h

NAME                                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/telegraf-ingress-deployment   1/1     1            1           40h

NAME                                                    DESIRED   CURRENT   READY   AGE
replicaset.apps/telegraf-ingress-deployment-ddfc8ff47   1         1         1       40h

NAME                        READY   AGE
statefulset.apps/openbmpd   1/1     40h
```

4. Configure routers in the network to stream telemetry and BMP data to the Jalapeno cluster. The MDT port is 32400 and the BMP port is 30555.

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
       host <server-ip> port 30555
       description jalapeno OpenBMP
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

5. If using Microk8s, navigate to the dashboard and check invidual services as appropriate.
```
http://<server-ip>:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/
```

## Destroying Jalapeno

Jalapeno can also be destroyed using the script.

1. Use the `destroy_jalapeno.sh` script. Will remove both namespaces jalapeno and jalapeno-collectors and all associated services/pods/deployments/etc. and it will remove all the persistent volumes associated with kafka and arangodb.

   ```shell
   destroy_jalapeno.sh kubectl
   ```

### More info on Jalapeno components:

* [MicroK8s_installation.md](docs/MicroK8s_installation.md)

* [Link-State processor](docs/link-state)

* [L3VPN processor](docs/l3vpn)

* [EPE processor](docs/epe)

* [BMP](docs/BMP.md) - coming soon

* [Kafka](docs/Kafka.md) - coming soon

* [Topology processor](docs/Topology_processor.md) - coming soon

* [Arango GraphDB](docs/Arango-GraphDB.md) - coming soon

* [Influx TSDB](docs/Influx-TSDB.md) - coming soon

* [Network-performance processors](docs/perf) - coming soon
