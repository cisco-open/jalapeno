### The following instructions will create a basic virtual network topology as shown here:
![jalapeno_base_testbed](https://wwwin-github.cisco.com/spa-ie/jalapeno/blob/brmcdoug/docs/jalapeno_base_testbed.png "jalapeno-base-testbed")

The base topology allows one to create and test Jalapeno virtual topology use cases including:
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
* Create an /opt/images/jalapeno directory
* Copy xrv9k, and any other image files to /opt/images/jalapeno
* Copy the xml files from [libvirt](libvirt) to a directory of your choice
Copy 

#### 4. Run testbed_setup.sh shell script
[testbed_virtual_network_setup.sh[(testbed_virtual_network_setup.sh)

#### 4. Define and launch VMs
    virsh define r00.xml
    virsh define r01.xml
    virsh define r02.xml
    virsh define r05.xml
    virsh define r06.xml

    virsh start r00
    virsh start r01
    virsh start r02
    virsh start r05
    virsh start r06

#### 5. Router console access:
    r00 -> telnet localhost 20000
    r05 -> telnet localhost 20050

#### 6. Once the virtual routers are up and running:

[xr-router-config.md](xr-router-config.md)
