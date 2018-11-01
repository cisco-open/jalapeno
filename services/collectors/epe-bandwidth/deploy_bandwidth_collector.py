#!/usr/bin/python
"""Deploy Bandwidth Collector Service in OpenShift.
Assumes the yaml files in the directory are
properly configured for the current environment.
"""
import yaml, urllib3, subprocess
import openshift_deployer
from openshift import client as openshift_client, config
from kubernetes import client as kube_client, config as kube_config

def main():
    """Load APIs and deploy Bandwidth components using associated yaml files."""
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    config.load_kube_config()
    deployment_api = openshift_client.AppsV1Api()

    openshift_project = "voltron"
    #openshift_deployer.deploy_deployment(deployment_api, "paths_collector_dp.yaml", openshift_project)
    subprocess.call(["oc", "apply", "-f", "bandwidth_collector_dp.yaml"])


if __name__ == '__main__':
    main()
