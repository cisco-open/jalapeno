### Some reference CLI commands for users who deploy Jalapeno on the Openshift k8s orchestrator:

```bash
oc login <server ip_address>    # log in to your cluster
oc project <project_name>       # enter your project
oc get pods                     # get running pods 
oc get pv                       # get persistent-volumes
oc logs <pod_name>              # get logs from specified pod
oc exec -it <pod_name> /bin/sh  # enter bash in specified pod
oc apply -f <directory_name>/.  # deploy any YAML files configured in specified directory into the OpenShift cluster
```
