# Jalapeno Installation
Jalapeno has been primarily developed, tested, and operated on Ubuntu 18.04 (bare-metal or VM), and Google Kubernetes Engine. Recommended VM sizing for a test lab is 4 vCPU, 16GB memory, and 50G of disk.  If deploying in production or an test environment with large table sizes (full Internet table, 100k + internal or vpn prefixes), then we recommend a bare metal K8s cluster with two or more nodes.

Users who do not have a full Kubernetes or GKE deployment can get up and running quite quickly with Microk8s [Installing K8s](docs/K8s_installation.md)

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
NAME                                              READY   STATUS    RESTARTS   AGE
pod/arangodb-0                                    1/1     Running   0          3m12s
pod/grafana-deployment-7d4dd466f5-d8lm5           1/1     Running   0          3m9s
pod/influxdb-0                                    1/1     Running   0          3m11s
pod/kafka-0                                       1/1     Running   0          3m13s
pod/telegraf-egress-deployment-77d475d8d8-xvhtp   1/1     Running   2          3m5s
pod/topology-75958bdf79-9n8tj                     1/1     Running   2          2m50s
pod/zookeeper-0                                   1/1     Running   0          3m14s

NAME                  TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
service/arango-np     NodePort    10.106.206.60    <none>        8529:30852/TCP               3m12s
service/arangodb      ClusterIP   10.96.17.131     <none>        8529/TCP                     3m13s
service/broker        ClusterIP   10.96.137.177    <none>        9092/TCP                     3m14s
service/grafana       ClusterIP   10.96.248.207    <none>        3000/TCP                     3m8s
service/grafana-np    NodePort    10.102.250.199   <none>        3000:30300/TCP               3m7s
service/influxdb      ClusterIP   10.105.37.120    <none>        8086/TCP                     3m11s
service/influxdb-np   NodePort    10.108.158.189   <none>        8086:30308/TCP               3m10s
service/kafka         NodePort    10.105.91.121    <none>        9092:30092/TCP               3m13s
service/zookeeper     ClusterIP   10.100.169.211   <none>        2888/TCP,3888/TCP,2181/TCP   3m14s

NAME                                         READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/grafana-deployment           1/1     1            1           3m9s
deployment.apps/telegraf-egress-deployment   1/1     1            1           3m5s
deployment.apps/topology                     1/1     1            1           2m50s

NAME                                                    DESIRED   CURRENT   READY   AGE
replicaset.apps/grafana-deployment-7d4dd466f5           1         1         1       3m9s
replicaset.apps/telegraf-egress-deployment-77d475d8d8   1         1         1       3m5s
replicaset.apps/topology-75958bdf79                     1         1         1       2m50s

NAME                         READY   AGE
statefulset.apps/arangodb    1/1     3m13s
statefulset.apps/influxdb    1/1     3m11s
statefulset.apps/kafka       1/1     3m13s
statefulset.apps/zookeeper   1/1     3m14s
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


