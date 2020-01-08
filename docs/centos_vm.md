# Centos VM packaging
As of January 2020 Voltron/Jalapeno's Openshift cluster is packaged on a Centos VM.  The VM may be launched via standard virsh (see os_base1.xml) 
Once launched, the user will need to modify a few openshift config files on the VM, and add iptables rules on the server host to enable port forwarding
to Openshift and Voltron/Jalapeno's services that come with UI's.

## Procedure:
```
1. copy centos VM and supplemental storage qcow2's to the VM image repository on your server host
openshift1.qcow2
openshift1-vdb.qcow2

2. copy os_base1.xml to your server host and define the VM:

virsh define os_base1.xml

3. Create the following iptables port forwarding entries on your server host

```
