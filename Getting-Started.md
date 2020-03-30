# Jalapeno Installation Guide

### Installing Jalapeno

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

### Pre-Installation

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
