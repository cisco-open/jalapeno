# OpenShift CLI
Almost all of Jalapeno's infrastructure and services are deployed on an OpenShift cluster.  
Thus, the OpenShift CLI client must be installed, as 'oc' commands are used through the Jalapeno deployment process.

## Installation
To install on CentosKVM, run:
```bash
# official release
wget https://github.com/openshift/origin/releases/download/v1.5.1/openshift-origin-client-tools-v1.5.1-7b451fc-linux-64bit.tar.gz
tar -xzf openshift-origin-client-tools-v1.5.1-7b451fc-linux-64bit.tar.gz
# install
sudo mv oc /bin/
```

## Command Reference
```bash
oc login <server ip_address>    # log in to your cluster
oc project <project_name>       # enter your project
oc get pods                     # get running pods 
oc get pv                       # get persistent-volumes
oc logs <pod_name>              # get logs from specified pod
oc exec -it <pod_name> /bin/sh  # enter bash in specified pod
oc apply -f <directory_name>/.  # deploy any YAML files configured in specified directory into the OpenShift cluster
```
