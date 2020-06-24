# MicroK8s

If you already have a Kubernetes environment deployed you may skip this document and move on to:

[Getting-Started.md](../Getting-Started.md)

If you do not have an existing Kubernetes environment, then the following instructions may be used to install a single-node Microk8s cluster.  Microk8s is great for development and is contained to a single server.

### Installing MicroK8s on Ubuntu

We will leverge this guide: [https://tutorials.ubuntu.com/tutorial/install-a-local-kubernetes-with-microk8s#0](https://tutorials.ubuntu.com/tutorial/install-a-local-kubernetes-with-microk8s#0)

1. Ensure that `snap` is installed using `snap version` . If snap is not installed `sudo apt install snapd`.

2. Install MicroK8s: `sudo snap install microk8s --classic --channel=1.17/stable`

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
8. Edit dashboard yaml, change ClusterIP to NodePort:
```
kubectl -n kube-system edit service kubernetes-dashboard
```
```
spec:
  clusterIP: 10.152.183.93
  externalTrafficPolicy: Cluster
  ports:
  - nodePort: 31444
    port: 443
    protocol: TCP
    targetPort: 8443
  selector:
    k8s-app: kubernetes-dashboard
  sessionAffinity: None
  type: NodePort
  ```

9. Enable skip for login token (only way over http proxy)

   1. `microk8s.kubectl -n kube-system edit deploy kubernetes-dashboard -o yaml`

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

10. Enable Kubernetes proxy in the background to access dashboard from your browser:
```
microk8s.kubectl proxy --accept-hosts=.* --address=0.0.0.0 &
```

11. Access Kubernetes Dashboard: 

```
To get Dashboard port number:

microk8s kubectl -n kube-system get service kubernetes-dashboard

Then

https://<server_ip>:<port_number>/pod?namespace=_all
```

12. Optional - if you wish to just use 'kubectl' without needing to type 'microk8s kubectl <command>'
```
sudo chown -f -R $USER ~/.kube
sudo snap install kubectl --classic
microk8s kubectl config view --raw > $HOME/.kube/config
```

### You may now proceed to install/deploy Jalapeno on your Microk8s cluster:

[Getting-Started.md](../Getting-Started.md)

## Removing Microk8s

Simply shutdown the cluster then remove Microk8s with snap:
```
microk8s stop
sudo snap remove microk8s
# verify the microk8s directory has been removed:
ls /var/snap 
```


