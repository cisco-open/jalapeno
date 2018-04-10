import yaml, urllib3
import openshift_deployer
from openshift import client as openshift_client, config
from kubernetes import client as kube_client, config as kube_config

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

config.load_kube_config()
deployment_api = openshift_client.AppsV1Api()
stateful_set_api = openshift_client.AppsV1Api()
service_api = kube_client.CoreV1Api()
route_api = openshift_client.OapiApi()
pod_api = kube_client.CoreV1Api()
persistent_volume_api = kube_client.CoreV1Api()
persistent_volume_claim_api = kube_client.CoreV1Api()
config_map_api = kube_client.CoreV1Api()

openshift_project = "voltron"
# openshift_deployer.deploy_deployment(deployment_api, "deployment.yaml", openshift_project)
# openshift_deployer.deploy_stateful_set(stateful_set_api, "stateful_set.yaml", openshift_project)
openshift_deployer.deploy_service(service_api, "service.yaml", openshift_project)
openshift_deployer.deploy_route(route_api, "route.yaml", openshift_project)
# openshift_deployer.deploy_pod(pod_api, "pod.yaml", openshift_project)
# openshift_deployer.deploy_persistent_volume(persistent_volume_api, "persistent_volume.yaml", openshift_project)
# openshift_deployer.deploy_persistent_volume_claim(persistent_volume_claim_api, "persistent_volume_claim.yaml", openshift_project)
# openshift_deployer.deploy_config_map(config_map_api, "config-map.yaml", openshift_project)


