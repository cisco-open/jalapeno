# MicroK8s

MicroK8s is a kuberenetes single node cluster. It is great for development and is contained to a single server.

## Getting Started on Ubuntu

We will leverge this guide: [https://tutorials.ubuntu.com/tutorial/install-a-local-kubernetes-with-microk8s#0](https://tutorials.ubuntu.com/tutorial/install-a-local-kubernetes-with-microk8s#0)

1. Ensure that `snap` is installed using `snap version` . If snap is not installed `sudo apt install snapd`.

2. Install MicroK8s: `sudo snap install microk8s --classic --edge`

3. Add user into microk8s group. `sudo usermod -a -G microk8s $USER` (You will need to reinstantiate shell for this to take effect)

4. Configure firewall to allow pod to pod communication (Note: If ufw is not installed use `sudo apt install ufw`):

   ```bash
   sudo apt install ufw
   sudo ufw allow in on cni0 && sudo ufw allow out on cni0
   sudo ufw default allow routed
   ```
   
5. If you are in a Cisco Lab w/ Proxies, configure the proxies by adding `HTTPS_PROXY=http://proxy.esl.cisco.com:8080` to `/var/snap/microk8s/current/args/containerd-env`.

    Then restart containerd `sudo systemctl restart snap.microk8s.daemon-containerd.service`

6. Enable dashboard helm and dns: `microk8s.enable dashboard dns`

7. Check if services are up: `microk8s.kubectl get all --all-namespaces`. The output should look like below. Take note of the columns `READY` and `STATUS`.

   ```shell
   kkumara3@ie-dev1:~$ microk8s.kubectl get all --all-namespaces
   NAMESPACE     NAME                                                  READY   STATUS    RESTARTS   AGE
   kube-system   pod/coredns-9b8997588-qxxpf                           1/1     Running   0          2d4h
   kube-system   pod/dashboard-metrics-scraper-687667bb6c-k85m7        1/1     Running   0          2d4h
   kube-system   pod/heapster-v1.5.2-5c58f64f8b-c8h4x                  4/4     Running   0          2d4h
   kube-system   pod/kubernetes-dashboard-5c848cc544-xrtgw             1/1     Running   0          2d4h
   kube-system   pod/monitoring-influxdb-grafana-v4-6d599df6bf-2c5g5   2/2     Running   0          2d4h

   NAMESPACE     NAME                                TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                  AGE
   default       service/kubernetes                  ClusterIP   10.152.183.1     <none>        443/TCP                  3d3h
   kube-system   service/dashboard-metrics-scraper   ClusterIP   10.152.183.131   <none>        8000/TCP                 2d4h
   kube-system   service/heapster                    ClusterIP   10.152.183.170   <none>        80/TCP                   2d4h
   kube-system   service/kube-dns                    ClusterIP   10.152.183.10    <none>        53/UDP,53/TCP,9153/TCP   2d4h
   kube-system   service/kubernetes-dashboard        ClusterIP   10.152.183.242   <none>        443/TCP                  2d4h
   kube-system   service/monitoring-grafana          ClusterIP   10.152.183.133   <none>        80/TCP                   2d4h
   kube-system   service/monitoring-influxdb         ClusterIP   10.152.183.87    <none>        8083/TCP,8086/TCP        2d4h

   NAMESPACE     NAME                                             READY   UP-TO-DATE   AVAILABLE   AGE
   kube-system   deployment.apps/coredns                          1/1     1            1           2d4h
   kube-system   deployment.apps/dashboard-metrics-scraper        1/1     1            1           2d4h
   kube-system   deployment.apps/heapster-v1.5.2                  1/1     1            1           2d4h
   kube-system   deployment.apps/kubernetes-dashboard             1/1     1            1           2d4h
   kube-system   deployment.apps/monitoring-influxdb-grafana-v4   1/1     1            1           2d4h

   NAMESPACE     NAME                                                        DESIRED   CURRENT   READY   AGE
   kube-system   replicaset.apps/coredns-9b8997588                           1         1         1       2d4h
   kube-system   replicaset.apps/dashboard-metrics-scraper-687667bb6c        1         1         1       2d4h
   kube-system   replicaset.apps/heapster-v1.5.2-5c58f64f8b                  1         1         1       2d4h
   kube-system   replicaset.apps/kubernetes-dashboard-5c848cc544             1         1         1       2d4h
   kube-system   replicaset.apps/monitoring-influxdb-grafana-v4-6d599df6bf   1         1         1       2d4h
   ```

8. Enable skip for login token (only way over http proxy)

   1. `sudo microk8s.kubectl -n kube-system edit deploy kubernetes-dashboard -o yaml`

   2. Add the `-enable-skip-login` flag to deployment's specs

      ```yaml
      spec:
        progressDeadlineSeconds: 600
        replicas: 1
        revisionHistoryLimit: 10
        selector:
          matchLabels:
            k8s-app: kubernetes-dashboard
        strategy:
          rollingUpdate:
            maxSurge: 25%
            maxUnavailable: 25%
          type: RollingUpdate
        template:
          metadata:
            creationTimestamp: null
            labels:
              k8s-app: kubernetes-dashboard
          spec:
            containers:
            - args:
              - --auto-generate-certificates
              - --namespace=kube-system
              - --enable-skip-login
      ```

9. Enable Kubernetes proxy in the background to access dashboard from your browser: `microk8s.kubectl proxy --accept-hosts=.* --address=0.0.0.0 &`

10. Access Kubernetes Dashboard: `http://<server-ip>:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/`

## Deploying Jalapeno

Jalapeno is very easy to deploy in this single cluster environment.

Note: prior to deploying, we recommend setting your Internal BGP ASN, and optionally, the ASNs of any direct or transit BGP peers you wish to track.  These settings are found in:

https://github.com/cisco-ie/jalapeno/blob/master/processors/topology/topology_dp.yaml

Example:
```
        args:
          - --asn
          - "100000"
          - --transit-provider-asns
          - "7200 7600"
          - --direct-peer-asns
          - "7100"
```

1. Clone repo and `cd` into folder: `git clone <repo> && cd jalapeno`

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

3. Use the `deploy_jalapeno.sh` script. This will start the collectors and all jalapeno infra and services on the single node.

   ```bash
   deploy_jalapeno.sh microk8s.kubectl
   ```

4. Check that all services are up using: `microk8s.kubectl get all --all-namespaces`

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
