#!/usr/bin/python
"""Deploy OpenBMPD in OpenShift.
Assumes the yaml files in the directory are
properly configured for the current environment.
"""
import yaml, urllib3, subprocess
import openshift_deployer
from openshift import client as openshift_client, config
from kubernetes import client as kube_client, config as kube_config

def main():
    """Load APIs and deploy OpenBMPD components using associated yaml files."""
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    config.load_kube_config()
    stateful_set_api = openshift_client.AppsV1Api()
    service_api = kube_client.CoreV1Api()
    persistent_volume_api = kube_client.CoreV1Api()

    openshift_project = "voltron"
#    openshift_deployer.deploy_service(service_api, "openbmpd_svc.yaml", openshift_project)
 #   openshift_deployer.deploy_service(service_api, "openbmpd_svc_np.yaml", openshift_project)
    #openshift_deployer.deploy_stateful_set(stateful_set_api, "openbmpd_ss.yaml", openshift_project)
    subprocess.call(["oc", "apply", "-f", "openbmpd_ss.yaml"])

if __name__ == '__main__':
    main()
