# Openshift
```
git clone https://github.com/openshift/openshift-ansible.git
cd openshift-ansible && git checkout release-1.5
```

Run the playbook with the hosts file provided, obviously modified to reflect your environment

# Cisco-Openshift

<Omar Fill in details. Please replace the above description>

## Framework
The framework has a few components, ArangoDB, Kafka, Framework and Topology. They
should be deployed in this order. The ArangoDB pods and Kafka pods need persistent
storage, which means the Openshift cluster must support some type of persistence.
The naive way is to use HostPath, this will only work in a single maste/node environment.
If the cluster is larger you must use a distributed safe storage method, such as
an NFS server or GlusterFS. Once the storage is set up you must set up add volumes.
We could not figure out how to do this via the UI, but it works using the CLI `oc apply -f`.

## Check for hostname / IP consistency
We've committed our last deployment to RTP, you will have to change your configuration.
Be aware of the `hostnames`, `IPs`, `persistent volumes`

## Install
```
oc apply -f kafka/.
oc apply -f arango/.
oc apply -f framework/.
oc apply -f topology/.
oc apply -f latency/.
oc apply -f responder/.
```
