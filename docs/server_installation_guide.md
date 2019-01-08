###1. Requirements: 
    * ubuntu 18.04, minimum 16 vCPU, 96GB memory, 200GB disk

###2. Required packages:
    * apt-get install openssh-server lxc lxd-client qemu qemu-kvm libvirt-bin openvswitch-switch python-pip git
    * optional: apt-get install virt-manager kafkacat

###3. Copy voltron tar files to /opt/vlt/

###4. Get the server's outside eth interface name.  Edit vlt_startup.sh and replace <server outside interface> with the interface name. Uncomment the iptables masquerade line

###5. cd into /opt/vlt and run vlt_startup.sh

###6. Once the script is complete, check the status of routers/vms, verify routes, etc:
    Example XR router console access:
    R00 -> telnet localhost 20000
    R05 -> telnet localhost 20050
    Example CSR router console access:
    sudo virsh console r71

###7. Modify openshift's public or reachable IP address:
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

###8. Once the virtual network is up and routing is established, ssh to the openshift VM and run voltron k8s deployment script:
        ssh centos@10.0.250.2, pw cisco
        cd ~/voltron
        ./deploy_voltron.sh 

###9. Verify openshift voltron project is up and pods have all started
        https://<server_ip>:8443/console/project/voltron/browse/pods

###10. Validate voltron data:
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

##Optional:

###11. Add latencies to the topology. Examples:
sudo tc qdisc add dev r71ge1 root netem delay 120000
sudo tc qdisc add dev r71ge2 root netem delay 150000
sudo tc qdisc add dev r72ge1 root netem delay 180000
sudo tc qdisc add dev r72ge2 root netem delay 210000