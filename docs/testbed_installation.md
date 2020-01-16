### The following instructions will create a basic virtual network topology as shown here:
![voltron_base_testbed](https://wwwin-github.cisco.com/spa-ie/voltron/blob/brmcdoug/docs/voltron_base_testbed.png "voltron-base-testbed")

The base topology allows one to create and test Voltron virtual topology use cases including:
* Internal traffic engineering - steering traffic over a non-IGP best path toward an internal or external destination
* Egress Peer Engineering - steering traffic out a non-BGP best path toward an external destination
* VPN Overlays - creation of VPN tunnels using SR or SRv6 encapsulations
* Service Chaining - creation of overlay tunnels linking sources and destinations with middle-point services in between
** see also: https://tools.ietf.org/html/draft-ietf-spring-sr-service-programming-01

#### 1. Server Requirements: 
    * ubuntu 18.04, minimum 16 vCPU, 96GB memory, 200GB disk

#### 2. Required packages:
    * apt-get install openssh-server lxc lxd-client qemu qemu-kvm libvirt-bin openvswitch-switch python-pip git
    * optional: apt-get install virt-manager kafkacat

#### 3. Copy VM files over
* Create an /opt/images/voltron directory
* Copy the Openshift Centos VM (os_base1) qcow2, xrv9k, and any other image files to /opt/images/voltron
* Copy the libvirt xml files from https://wwwin-github.cisco.com/spa-ie/voltron/edit/brmcdoug/docs/libvirt/ to a directory of your choice
Copy 

#### 4. Run testbed_setup.sh shell script
    https://wwwin-github.cisco.com/spa-ie/jalapeno/blob/master/docs/testbed_virtual_network_setup.sh

#### 4. Define and launch VMs
    virsh define r00.xml
    virsh define r01.xml
    virsh define r02.xml
    virsh define r05.xml
    virsh define r06.xml
    virsh define r71.xml
    virsh define r72.xml
    virsh define os_base1.xml

    virsh start r00
    virsh start r01
    virsh start r02
    virsh start r05
    virsh start r06
    virsh start r71
    virsh start r72
    virsh start os_base1

#### 5. Router console access:
    r00 -> telnet localhost 20000
    r05 -> telnet localhost 20050
    Example CSR router console access:
    sudo virsh console r71

#### 6. Modify openshift's public or reachable IP address:
(more info on installing and configuing the Openshif centos VM at https://wwwin-github.cisco.com/spa-ie/voltron/blob/brmcdoug/docs/centos_vm.md )

    1. virsh console os_base1
    2. shutdown openshift
        sudo systemctl stop origin-node
        sudo systemctl stop origin-node-dep
        sudo systemctl stop origin-master

    3. edit these files - replace all 10.200.99.x ip addresses with your server's outside IP
        /etc/origin/master/openshift-master.kubeconfig
        /etc/origin/master/admin.kubeconfig
        /etc/origin/master/master-config.yaml

    4. Bypass any proxies:
        /etc/environment
            export no_proxy=localhost,127.0.0.1,os_base1,10.200.99.4

    5. restart openshift
        sudo systemctl start origin-master
        sudo systemctl start origin-node
        sudo systemctl start origin-node-dep

#### 7. Once the virtual network is up and routing is established, ssh to the openshift VM and run voltron k8s deployment script:
        ssh centos@10.0.250.2, pw cisco
        cd ~/voltron
        ./deploy_voltron.sh 

#### 8. Verify openshift voltron project is up and pods have all started
        https://<server_ip>:8443/console/project/voltron/browse/pods

#### 9. Validate voltron data:
    1. OpenBMP:
     docker exec -it openbmp_collector bash
      more /var/log/openbmpd.log

    2. Kafka: validate bmp data is reaching kafka - login to zookeeper, then:
     cd /opt/kafka/bin
     Example commands:
    ./kafka-topics.sh --zookeeper zookeeper.voltron.svc:2181 --list
     ./kafka-console-consumer.sh --zookeeper zookeeper.voltron.svc:2181 --topic openbmp.parsed.ls_link --from-beginning
     ./kafka-console-consumer.sh --zookeeper zookeeper.voltron.svc:2181 --topic openbmp.parsed.ls_node --from-beginning
     ./kafka-console-consumer.sh --zookeeper zookeeper.voltron.svc:2181 --topic openbmp.parsed.ls_prefix --from-beginning
     ./kafka-console-consumer.sh --zookeeper zookeeper.voltron.svc:2181 --topic openbmp.parsed.unicast_prefix --from-beginning

    3. Arango:
     http://<server_ip>:30852

### Optional:

#### 11. Add latencies to the topology. Examples:
    sudo tc qdisc add dev r71ge1 root netem delay 120000 <br>
    sudo tc qdisc add dev r71ge2 root netem delay 150000 <br>
    sudo tc qdisc add dev r72ge1 root netem delay 180000 <br>
    sudo tc qdisc add dev r72ge2 root netem delay 210000 <br>
