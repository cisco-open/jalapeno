#!/usr/bin/python
"""Deploy Grafana in OpenShift.
Assumes the yaml files in the directory are
properly configured for the current environment.
"""
import yaml, urllib3, subprocess
import openshift_deployer
from openshift import client as openshift_client, config
from kubernetes import client as kube_client, config as kube_config

def main():
    """Load APIs and deploy Grafana components using associated yaml files."""
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    config.load_kube_config()
    deployment_api = openshift_client.AppsV1Api()
    service_api = kube_client.CoreV1Api()
    config_map_api = kube_client.CoreV1Api()

    openshift_project = "voltron"
    # openshift_deployer.deploy_deployment(deployment_api, "grafana_dp.yaml", openshift_project)
    subprocess.call(["oc", "apply", "-f", "grafana_dp.yaml"])
    openshift_deployer.deploy_service(service_api, "grafana_svc.yaml", openshift_project)
    openshift_deployer.deploy_config_map(config_map_api, "grafana_cfg.yaml", openshift_project)

if __name__ == '__main__':
    main()
