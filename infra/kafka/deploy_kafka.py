#!/usr/bin/python
"""Deploy Kafka in OpenShift.
Assumes the yaml files in the directory are
properly configured for the current environment.
"""
import yaml, urllib3, subprocess
import openshift_deployer
from openshift import client as openshift_client, config
from kubernetes import client as kube_client, config as kube_config

def main():
    """Load APIs and deploy Kafka components using associated yaml files."""
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    config.load_kube_config()
    config_map_api = kube_client.CoreV1Api()
    service_api = kube_client.CoreV1Api()
    persistent_volume_api = kube_client.CoreV1Api()
    #stateful_set_api = openshift_client.CoreV1Beta1Api()

    openshift_project = "voltron"
    openshift_deployer.deploy_config_map(config_map_api, "zookeeper_cfg.yaml", openshift_project)
    openshift_deployer.deploy_service(service_api, "zookeeper_svc.yaml", openshift_project)
    openshift_deployer.deploy_service(service_api, "zookeeper_svc_np.yaml", openshift_project)
    openshift_deployer.deploy_persistent_volume(persistent_volume_api, "zookeeper_pv.yaml", openshift_project)
    #openshift_deployer.deploy_stateful_set(stateful_set_api, "zookeeper_ss.yaml", openshift_project)
    subprocess.call(["oc", "apply", "-f", "zookeeper_ss.yaml"])
    openshift_deployer.deploy_config_map(config_map_api, "broker_cfg.yaml", openshift_project)
    openshift_deployer.deploy_persistent_volume(persistent_volume_api, "kafka_pv.yaml", openshift_project)
    openshift_deployer.deploy_service(service_api, "broker_svc.yaml", openshift_project)
    openshift_deployer.deploy_service(service_api, "kafka_svc_np.yaml", openshift_project)
    #openshift_deployer.deploy_stateful_set(stateful_set_api, "kafka_ss.yaml", openshift_project)
    subprocess.call(["oc", "apply", "-f", "kafka_ss.yaml"])


if __name__ == '__main__':
    main()
