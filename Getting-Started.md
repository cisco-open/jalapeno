# Jalapeno Installation Guide
The following instructions use command line examples when installing/deploying on a Microk8s cluster

### Installing Jalapeno

1. Clone this repo and `cd` into the folder: `git clone <repo> && cd jalapeno`

2. Ensure that you have a Docker login set up via `sudo docker login` command that has access to docker.io/iejalapeno. **Note: You need docker installed for this step**

   ```bash
   $ cat $HOME/.docker/config.json
   {
    "auths": {
      "https://index.docker.io/v1/": {
        "auth": "c2trdW1hcmF2Zqweqwea2FyNzYxNw=="
      }
    },
    "HttpHeaders": {
      "User-Agent": "Docker-Client/19.03.5 (linux)"
   }
   ```

### Pre-Deployment

Jalapeno's Topology processing makes a distinction between Internal topology: the nodes, links, prefixes, ASNs, etc, that make up the internal network; and External topology: the Internet, or other ASNs that we connect to but are not under our administrative control. Thus, prior to deploying, we recommend configuring the Topology processor to identify your Internal BGP ASN(s), and optionally, the ASNs of any direct or transit BGP peers you wish to track.  These settings are found in:

https://github.com/cisco-ie/jalapeno/blob/master/processors/topology/topology_dp.yaml

Note, private BGP ASNs are accounted for as Internal by default.  We may include a knob in the future which allows private ASNs to be considered External if needed.

Example from topology_dp.yaml:
```
        args:
          - --asn
          - "109 36692 13445"
          - --transit-provider-asns
          - "3356 2914"
          - --direct-peer-asns
          - "2906 8075"
```

3. Use the `deploy_jalapeno.sh` script. This will start the collectors and all jalapeno infra and services on the single node.

   ```bash
   deploy_jalapeno.sh microk8s.kubectl
   ```

4. Check that all containers are up using: `microk8s.kubectl get all --all-namespaces` or on a per-namespace basis:
```
microk8s.kubectl get all -n jalapeno
microk8s.kubectl get all -n jalapeno-collectors
```
Output
```
microk8s.kubectl get all -n jalapeno

NAME                                              READY   STATUS             RESTARTS   AGE
pod/arangodb-0                                    1/1     Running            1          40h
pod/grafana-deployment-5f44494444-d8smc           1/1     Running            1          40h
pod/influxdb-0                                    1/1     Running            1          40h
pod/kafka-0                                       1/1     Running            2          40h
pod/l3vpn-processor-574d6c6d78-xpqd6              1/1     Running            70         40h
pod/ls-performance-processor-5b4865464b-zck7w     1/1     Running            36         35h
pod/ls-processor-667b548b46-ch4pp                 1/1     Running            20         34h
pod/telegraf-egress-deployment-544b5c757c-rkvg5   1/1     Running            3          39h
pod/topology-75bfcb977-c85z9                      1/1     Running            3          40h
pod/zookeeper-0                                   1/1     Running            1          40h

NAME                  TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
service/arango-np     NodePort    10.152.183.22    <none>        8529:30852/TCP               40h
service/arangodb      ClusterIP   10.152.183.68    <none>        8529/TCP                     40h
service/broker        ClusterIP   10.152.183.206   <none>        9092/TCP                     40h
service/grafana       ClusterIP   10.152.183.176   <none>        3000/TCP                     40h
service/grafana-np    NodePort    10.152.183.80    <none>        3000:30300/TCP               40h
service/influxdb      ClusterIP   10.152.183.81    <none>        8086/TCP                     40h
service/influxdb-np   NodePort    10.152.183.231   <none>        8086:30308/TCP               40h
service/kafka         NodePort    10.152.183.232   <none>        9092:30092/TCP               40h
service/zookeeper     ClusterIP   10.152.183.105   <none>        2888/TCP,3888/TCP,2181/TCP   40h

NAME                                         READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/grafana-deployment           1/1     1            1           40h
deployment.apps/l3vpn-processor              1/1     1            1           40h
deployment.apps/ls-performance-processor     1/1     1            1           40h
deployment.apps/ls-processor                 1/1     1            1           40h
deployment.apps/telegraf-egress-deployment   1/1     1            1           40h
deployment.apps/topology                     1/1     1            1           40h

NAME                                                    DESIRED   CURRENT   READY   AGE
replicaset.apps/grafana-deployment-5f44494444           1         1         1       40h
replicaset.apps/l3vpn-processor-574d6c6d78              1         1         1       40h
replicaset.apps/ls-performance-processor-5b4865464b     1         1         1       40h
replicaset.apps/ls-processor-56dbcbd8d7                 1         1         0       22h
replicaset.apps/ls-processor-667b548b46                 1         1         1       34h
replicaset.apps/ls-processor-7d6d4c7554                 0         0         0       40h
replicaset.apps/ls-processor-7db9dc445                  0         0         0       22h
replicaset.apps/telegraf-egress-deployment-544b5c757c   1         1         1       40h
replicaset.apps/topology-75bfcb977                      1         1         1       40h

NAME                         READY   AGE
statefulset.apps/arangodb    1/1     40h
statefulset.apps/influxdb    1/1     40h
statefulset.apps/kafka       1/1     40h
statefulset.apps/zookeeper   1/1     40h
```
Collectors
```
microk8s.kubectl get all -n jalapeno-collectors
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

5. Configure routers in the network to stream telemetry and BMP data to the Jalapeno cluster. The MDT port is 32400 and the BMP port is 30555.

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

6. If using Microk8s, navigate to the dashboard and check invidual services as appropriate.
```
http://<server-ip>:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/
```

## Destroying Jalapeno

Jalapeno can also be destroyed using the script.

1. Use the `destroy_jalapeno.sh` script. Will remove both namespaces jalapeno and jalapeno-collectors and all associated services/pods/deployments/etc. and it will remove all the persistent volumes associated with kafka and arangodb.

   ```shell
   destory_jalapeno.sh microk8s.kubectl
   ```

### More info:

* [MicroK8s_installation.md](MicroK8s_installation.md)

* [BMP](BMP.md)

* [Kafka](Kafka.md)

* [Topology_processor](Topology_processor.md)

* [Arango-GraphDB](Arango-GraphDB.md)

* [Influx-TSDB](Influx-TSDB.md)

* [Link-State_processor](Link-State_processor.md)

* [L3VPN_processor](L3VPN_processor.md)

* [EPE_processor](EPE_processor.md)

* [Network-performance_processors](Network-performance_processors.md)
