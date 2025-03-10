# Kubernetes

!!! info

    If you already have a Kubernetes environment deployed you may skip this document and move on to the [Jalapeno Install](jalapeno.md)

The following instructions present two options for installing Kubernetes:

1. [Kubernetes/Kubeadm](#kubernetes-install)
2. [Microk8s](#microk8s-install)

## Kubernetes Install

Setting up a K8s cluster is outside the scope of this documentation. Instead, begin with the following guides:

1. [Install Kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)

2. [Create a cluster](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/)

???+ note

    Ubuntu system pre-flight checks may error out with:

    `[WARNING IsDockerSystemdCheck]: detected "cgroupfs" as the Docker cgroup driver. The recommended driver is "systemd". Please follow the guide at https://kubernetes.io/docs/setup/cri/
    `
    

    To change Docker driver to systemd add this line to `/etc/docker/daemon.json`: 
    
    ```json
    {
        "exec-opts": ["native.cgroupdriver=systemd"]
    }
    ```
    

    Then restart docker with `sudo systemctl restart docker.service`

### Install Cilium CNI (Optional)

If desired, install the Cilium CNI using the instructions available [here](https://docs.cilium.io/en/stable/gettingstarted/k8s-install-default/).

## K8s Validation

Once installation is complete the cluster should look similar to the output below:

```{ .text .no-copy }
$ kubectl get all --all-namespaces

NAMESPACE             NAME                                               READY   STATUS    RESTARTS      AGE
jalapeno-collectors   pod/gobmp-54cc8cf9b9-4kfrj                         1/1     Running   1 (10d ago)   20d
jalapeno-collectors   pod/telegraf-ingress-deployment-77f868dd79-8fnz8   1/1     Running   2 (10d ago)   20d
jalapeno              pod/arangodb-0                                     1/1     Running   1 (10d ago)   20d
jalapeno              pod/grafana-deployment-58986bc44b-gpq7n            1/1     Running   1 (10d ago)   20d
jalapeno              pod/influxdb-0                                     1/1     Running   1 (10d ago)   20d
jalapeno              pod/kafka-0                                        1/1     Running   2 (10d ago)   20d
jalapeno              pod/lslinknode-edge-744bd66695-pzhk2               1/1     Running   6 (10d ago)   20d
jalapeno              pod/telegraf-egress-deployment-84448c9879-dw8mc    1/1     Running   5 (10d ago)   20d
jalapeno              pod/topology-665c776f84-l8896                      1/1     Running   0             9d
jalapeno              pod/zookeeper-0                                    1/1     Running   1 (10d ago)   20d
kube-system           pod/cilium-envoy-kqjmq                             1/1     Running   1 (10d ago)   20d
kube-system           pod/cilium-operator-54c7465577-fvcms               1/1     Running   1 (10d ago)   20d
kube-system           pod/cilium-trrw8                                   1/1     Running   1 (10d ago)   20d
kube-system           pod/coredns-7c65d6cfc9-fdzc4                       1/1     Running   1 (10d ago)   20d
kube-system           pod/coredns-7c65d6cfc9-v4mv9                       1/1     Running   1 (10d ago)   20d
kube-system           pod/etcd-jalapeno-host                             1/1     Running   1 (10d ago)   20d
kube-system           pod/kube-apiserver-jalapeno-host                   1/1     Running   1 (10d ago)   20d
kube-system           pod/kube-controller-manager-jalapeno-host          1/1     Running   1 (10d ago)   20d
kube-system           pod/kube-proxy-dts8w                               1/1     Running   1 (10d ago)   20d
kube-system           pod/kube-scheduler-jalapeno-host                   1/1     Running   1 (10d ago)   20d

NAMESPACE             NAME                          TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                          AGE
default               service/kubernetes            ClusterIP   10.96.0.1        <none>        443/TCP                          20d
jalapeno-collectors   service/gobmp                 NodePort    10.96.245.167    <none>        5000:30511/TCP,56767:30767/TCP   20d
jalapeno-collectors   service/telegraf-ingress-np   NodePort    10.97.247.231    <none>        57400:32400/TCP                  20d
jalapeno              service/arango-np             NodePort    10.111.203.73    <none>        8529:30852/TCP                   20d
jalapeno              service/arangodb              ClusterIP   10.99.179.251    <none>        8529/TCP                         20d
jalapeno              service/broker                ClusterIP   10.103.212.223   <none>        9092/TCP                         20d
jalapeno              service/grafana               ClusterIP   10.99.46.64      <none>        3000/TCP                         20d
jalapeno              service/grafana-np            NodePort    10.104.190.91    <none>        3000:30300/TCP                   20d
jalapeno              service/influxdb              ClusterIP   10.111.183.55    <none>        8086/TCP                         20d
jalapeno              service/influxdb-np           NodePort    10.97.60.195     <none>        8086:30308/TCP                   20d
jalapeno              service/kafka                 NodePort    10.97.226.142    <none>        9094:30092/TCP                   20d
jalapeno              service/zookeeper             ClusterIP   10.109.47.153    <none>        2888/TCP,3888/TCP,2181/TCP       20d
kube-system           service/cilium-envoy          ClusterIP   None             <none>        9964/TCP                         20d
kube-system           service/hubble-peer           ClusterIP   10.110.213.58    <none>        443/TCP                          20d
kube-system           service/kube-dns              ClusterIP   10.96.0.10       <none>        53/UDP,53/TCP,9153/TCP           20d

NAMESPACE     NAME                          DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR            AGE
kube-system   daemonset.apps/cilium         1         1         1       1            1           kubernetes.io/os=linux   20d
kube-system   daemonset.apps/cilium-envoy   1         1         1       1            1           kubernetes.io/os=linux   20d
kube-system   daemonset.apps/kube-proxy     1         1         1       1            1           kubernetes.io/os=linux   20d

NAMESPACE             NAME                                          READY   UP-TO-DATE   AVAILABLE   AGE
jalapeno-collectors   deployment.apps/gobmp                         1/1     1            1           20d
jalapeno-collectors   deployment.apps/telegraf-ingress-deployment   1/1     1            1           20d
jalapeno              deployment.apps/grafana-deployment            1/1     1            1           20d
jalapeno              deployment.apps/lslinknode-edge               1/1     1            1           20d
jalapeno              deployment.apps/telegraf-egress-deployment    1/1     1            1           20d
jalapeno              deployment.apps/topology                      1/1     1            1           20d
kube-system           deployment.apps/cilium-operator               1/1     1            1           20d
kube-system           deployment.apps/coredns                       2/2     2            2           20d

NAMESPACE             NAME                                                     DESIRED   CURRENT   READY   AGE
jalapeno-collectors   replicaset.apps/gobmp-54cc8cf9b9                         1         1         1       20d
jalapeno-collectors   replicaset.apps/telegraf-ingress-deployment-77f868dd79   1         1         1       20d
jalapeno              replicaset.apps/grafana-deployment-58986bc44b            1         1         1       20d
jalapeno              replicaset.apps/lslinknode-edge-744bd66695               1         1         1       20d
jalapeno              replicaset.apps/telegraf-egress-deployment-84448c9879    1         1         1       20d
jalapeno              replicaset.apps/topology-665c776f84                      1         1         1       20d
kube-system           replicaset.apps/cilium-operator-54c7465577               1         1         1       20d
kube-system           replicaset.apps/coredns-7c65d6cfc9                       2         2         2       20d

NAMESPACE   NAME                         READY   AGE
jalapeno    statefulset.apps/arangodb    1/1     20d
jalapeno    statefulset.apps/influxdb    1/1     20d
jalapeno    statefulset.apps/kafka       1/1     20d
jalapeno    statefulset.apps/zookeeper   1/1     20d
```

If everything looks good, move onto [Installing Jalepeno](jalapeno.md).

## MicroK8s Install

The following instructions may be used to install a single-node Microk8s cluster.  

### Installing MicroK8s on Ubuntu

The following documentation is based on [this guide](https://tutorials.ubuntu.com/tutorial/install-a-local-kubernetes-with-microk8s#0).

1. Ensure that `snap` is installed using `snap version`.
     - If necessary, install snap with `sudo apt install snapd` or see the instructions [here](https://snapcraft.io/docs/installing-snapd).

2. Install MicroK8s using `sudo snap install microk8s --classic`
     - Optionally specify a specific channel using the flag `--channel=1.17/stable`

3. Add user into microk8s group with `sudo usermod -a -G microk8s $USER`
     - You will need to reinstantiate shell for this to take effect

4. Configure firewall to allow pod to pod communication (Note: If ufw is not installed use `sudo apt install ufw`):

    ```bash
    sudo apt install ufw
    sudo ufw allow in on cni0 && sudo ufw allow out on cni0
    sudo ufw default allow routed
    ```

5. If your cluster is behind a proxy, add the following to `/var/snap/microk8s/current/args/containerd-env`:

    ```text
    HTTPS_PROXY=http://<your_proxy>
    NO_PROXY=10.0.0.0/8
    ```

    Then restart containerd with `sudo systemctl restart snap.microk8s.daemon-containerd.service`

6. Enable dashboard helm and dns: `microk8s.enable dashboard dns`

7. Check if services are up: `microk8s.kubectl get all --all-namespaces`.

    The output should look like below. Take note of the columns `READY` and `STATUS`.

    ```{ .text .no-copy }
    $ microk8s.kubectl get all --all-namespaces
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

8. Edit dashboard yaml config to change ClusterIP to NodePort:

    ```bash
    microk8s.kubectl -n kube-system edit service kubernetes-dashboard
    ```

    Changes should be similar to the following:

    ```yaml
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

    ```bash
    microk8s.kubectl proxy --accept-hosts=.* --address=0.0.0.0 &
    ```

11. Access Kubernetes Dashboard

    To get Dashboard port number:

    ```bash
    microk8s kubectl -n kube-system get service kubernetes-dashboard
    ```

    Then attempt to reach the web UI at: `https://<server_ip>:<port_number>/pod?namespace=_all`

12. (Recommended) Enable 'kubectl' without needing to type 'microk8s kubectl <command>'

    ```bash
    sudo chown -f -R $USER ~/.kube
    sudo snap install kubectl --classic
    microk8s kubectl config view --raw > $HOME/.kube/config
    ```

You may now proceed to [Install Jalapeno](jalapeno.md) on your Microk8s cluster

### Removing Microk8s

Simply shutdown the cluster then remove Microk8s with snap:

```bash
microk8s stop
sudo snap remove microk8s
# verify the microk8s directory has been removed:
ls /var/snap 
```
