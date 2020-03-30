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

Jalapeno's Topology processing makes a distinction between Internal (topology, nodes, links, prefixes, ASNs) and External. Thus, prior to deploying, we recommend configuring the Topology processor to identify your Internal BGP ASN(s), and optionally, the ASNs of any direct or transit BGP peers you wish to track.  These settings are found in:

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

4. Check that all containers are up using: `microk8s.kubectl get all --all-namespaces`

```
NAMESPACE             NAME                                                  READY   STATUS              RESTARTS   AGE
jalapeno-collectors   pod/openbmpd-0                                        1/1     Running             0          118s
jalapeno-collectors   pod/telegraf-ingress-deployment-ddfc8ff47-mfcsb       0/1     CrashLoopBackOff    3          2m1s
jalapeno              pod/arangodb-0                                        1/1     Running             0          2m3s
jalapeno              pod/grafana-deployment-5f44494444-d8smc               1/1     Running             0          2m2s
jalapeno              pod/influxdb-0                                        1/1     Running             0          2m3s
jalapeno              pod/kafka-0                                           1/1     Running             0          2m3s
jalapeno              pod/l3vpn-processor-574d6c6d78-xpqd6                  1/1     Running             0          117s
jalapeno              pod/ls-performance-processor-5b4865464b-slccm         1/1     Running             0          116s
jalapeno              pod/ls-processor-7d6d4c7554-zbkzw                     1/1     Running             1          118s
jalapeno              pod/telegraf-egress-deployment-544b5c757c-kfvq9       0/1     ContainerCreating   0          2m2s
jalapeno              pod/topology-75bfcb977-c85z9                          1/1     Running             1          118s
jalapeno              pod/zookeeper-0                                       1/1     Running             0          2m4s
kube-system           pod/coredns-7b67f9f8c-mg9qn                           1/1     Running             9          12d
kube-system           pod/dashboard-metrics-scraper-687667bb6c-6cspm        1/1     Running             9          12d
kube-system           pod/heapster-v1.5.2-5c58f64f8b-8v4jh                  4/4     Running             36         12d
kube-system           pod/ip-masq-agent-gbbtn                               1/1     Running             6          8d
kube-system           pod/kubernetes-dashboard-67464df9f8-qfccd             1/1     Running             9          12d
kube-system           pod/monitoring-influxdb-grafana-v4-6d599df6bf-896np   2/2     Running             18         12d

NAMESPACE             NAME                                TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
default               service/kubernetes                  ClusterIP   10.152.183.1     <none>        443/TCP                      12d
jalapeno-collectors   service/openbmpd-np                 NodePort    10.152.183.23    <none>        5000:30555/TCP               2m
jalapeno-collectors   service/telegraf-ingress-np         NodePort    10.152.183.168   <none>        57400:32400/TCP              2m1s
jalapeno              service/arango-np                   NodePort    10.152.183.22    <none>        8529:30852/TCP               2m3s
jalapeno              service/arangodb                    ClusterIP   10.152.183.68    <none>        8529/TCP                     2m3s
jalapeno              service/broker                      ClusterIP   10.152.183.206   <none>        9092/TCP                     2m4s
jalapeno              service/grafana                     ClusterIP   10.152.183.176   <none>        3000/TCP                     2m2s
jalapeno              service/grafana-np                  NodePort    10.152.183.80    <none>        3000:30300/TCP               2m2s
jalapeno              service/influxdb                    ClusterIP   10.152.183.81    <none>        8086/TCP                     2m3s
jalapeno              service/influxdb-np                 NodePort    10.152.183.231   <none>        8086:30308/TCP               2m3s
jalapeno              service/kafka                       NodePort    10.152.183.232   <none>        9092:30092/TCP               2m4s
jalapeno              service/zookeeper                   ClusterIP   10.152.183.105   <none>        2888/TCP,3888/TCP,2181/TCP   2m4s
kube-system           service/dashboard-metrics-scraper   ClusterIP   10.152.183.205   <none>        8000/TCP                     12d
kube-system           service/heapster                    ClusterIP   10.152.183.46    <none>        80/TCP                       12d
kube-system           service/kube-dns                    ClusterIP   10.152.183.10    <none>        53/UDP,53/TCP,9153/TCP       12d
kube-system           service/kubernetes-dashboard        ClusterIP   10.152.183.154   <none>        443/TCP                      12d
kube-system           service/monitoring-grafana          ClusterIP   10.152.183.174   <none>        80/TCP                       12d
kube-system           service/monitoring-influxdb         ClusterIP   10.152.183.63    <none>        8083/TCP,8086/TCP            12d

NAMESPACE     NAME                           DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
kube-system   daemonset.apps/ip-masq-agent   1         1         1       1            1           <none>          8d

NAMESPACE             NAME                                             READY   UP-TO-DATE   AVAILABLE   AGE
jalapeno-collectors   deployment.apps/telegraf-ingress-deployment      0/1     1            0           2m1s
jalapeno              deployment.apps/grafana-deployment               1/1     1            1           2m2s
jalapeno              deployment.apps/l3vpn-processor                  1/1     1            1           118s
jalapeno              deployment.apps/ls-performance-processor         1/1     1            1           117s
jalapeno              deployment.apps/ls-processor                     1/1     1            1           118s
jalapeno              deployment.apps/telegraf-egress-deployment       0/1     1            0           2m2s
jalapeno              deployment.apps/topology                         1/1     1            1           119s
kube-system           deployment.apps/coredns                          1/1     1            1           12d
kube-system           deployment.apps/dashboard-metrics-scraper        1/1     1            1           12d
kube-system           deployment.apps/heapster-v1.5.2                  1/1     1            1           12d
kube-system           deployment.apps/kubernetes-dashboard             1/1     1            1           12d
kube-system           deployment.apps/monitoring-influxdb-grafana-v4   1/1     1            1           12d

NAMESPACE             NAME                                                        DESIRED   CURRENT   READY   AGE
jalapeno-collectors   replicaset.apps/telegraf-ingress-deployment-ddfc8ff47       1         1         0       2m1s
jalapeno              replicaset.apps/grafana-deployment-5f44494444               1         1         1       2m2s
jalapeno              replicaset.apps/l3vpn-processor-574d6c6d78                  1         1         1       117s
jalapeno              replicaset.apps/ls-performance-processor-5b4865464b         1         1         1       117s
jalapeno              replicaset.apps/ls-processor-7d6d4c7554                     1         1         1       118s
jalapeno              replicaset.apps/telegraf-egress-deployment-544b5c757c       1         1         0       2m2s
jalapeno              replicaset.apps/topology-75bfcb977                          1         1         1       118s
kube-system           replicaset.apps/coredns-7b67f9f8c                           1         1         1       12d
kube-system           replicaset.apps/dashboard-metrics-scraper-687667bb6c        1         1         1       12d
kube-system           replicaset.apps/heapster-v1.5.2-5c58f64f8b                  1         1         1       12d
kube-system           replicaset.apps/kubernetes-dashboard-67464df9f8             1         1         1       12d
kube-system           replicaset.apps/kubernetes-dashboard-6797dfbbf              0         0         0       12d
kube-system           replicaset.apps/monitoring-influxdb-grafana-v4-6d599df6bf   1         1         1       12d

NAMESPACE             NAME                         READY   AGE
jalapeno-collectors   statefulset.apps/openbmpd    1/1     2m
jalapeno              statefulset.apps/arangodb    1/1     2m3s
jalapeno              statefulset.apps/influxdb    1/1     2m3s
jalapeno              statefulset.apps/kafka       1/1     2m4s
jalapeno              statefulset.apps/zookeeper   1/1     2m4s
```


5. Configure the routers to point towards the cluster. The MDT port is 32400 and the BMP port is 30555.

   1. Example destination group for MDT: **Note: you may need to set TPA mgmt**

      ```shell
       destination-group jalapeno
        address-family ipv4 <server-ip> port 32400
         encoding self-describing-gpb
         protocol grpc no-tls
        !
       !
      ```

   2. Example of BMP config:

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
      ```

6. Navigate to the dashboard and check invidual services as appropriate.

## Destroying Jalapeno

Jalapeno can also be destroyed using the script.

1. Use the `destroy_jalapeno.sh` script. Will remove both namespaces jalapeno and jalapeno-collectors and all associated services/pods/deployments/etc. and it will remove all the persistent volumes associated with kafka and arangodb.

   ```shell
   destory_jalapeno.sh microk8s.kubectl
   ```
