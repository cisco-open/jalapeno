import yaml

def load_yaml(yaml_file):
    with open(yaml_file, 'r') as file:
        data=file.read()
    yaml_data = yaml.load(data)
    return yaml_data

def deploy_deployment(deployment_api, yaml_file, openshift_project):
    print("Deploying Deployment for " + openshift_project)
    yaml_data = load_yaml(yaml_file)
    deployment_resp = deployment_api.create_namespaced_deployment(body=yaml_data, namespace=openshift_project)
    print(deployment_resp.metadata.self_link)
    print("Deployed Deployment for " + openshift_project)

def deploy_stateful_set(stateful_set_api, yaml_file, openshift_project):
    print("Deploying Stateful Set for " + openshift_project)
    yaml_data = load_yaml(yaml_file)
    stateful_set_resp = stateful_set_api.create_namespaced_stateful_set(body=yaml_data, namespace=openshift_project)
    print(stateful_set_resp.metadata.self_link)
    print("Deployed Stateful Set for " + openshift_project)

def deploy_service(service_api, yaml_file, openshift_project):
    print("Deploying Service for " + openshift_project)
    yaml_data = load_yaml(yaml_file)
    service_resp = service_api.create_namespaced_service(body=yaml_data, namespace=openshift_project)
    print(service_resp.metadata.self_link)
    print("Deployed Service for " + openshift_project)

def deploy_route(route_api, yaml_file, openshift_project):
    print("Deploying Route for " + openshift_project)
    yaml_data = load_yaml(yaml_file)
    route_resp = route_api.create_namespaced_route(body=yaml_data, namespace=openshift_project)
    print(route_resp.metadata.self_link)
    print("Deployed Route for " + openshift_project)

def deploy_pod(pod_api, yaml_file, openshift_project):
    print("Deploying Pod for " + openshift_project)
    yaml_data = load_yaml(yaml_file)
    pod_resp = pod_api.create_namespaced_pod(body=yaml_data, namespace=openshift_project)
    print(pod_resp.metadata.self_link)
    print("Deployed Pod for " + openshift_project)

def deploy_persistent_volume(persistent_volume_api, yaml_file, openshift_project):
    print("Deploying Persistent Volume for " + openshift_project)
    yaml_data = load_yaml(yaml_file)
    persistent_volume_resp = persistent_volume_api.create_persistent_volume(body=yaml_data)
    print(persistent_volume_resp.metadata.self_link)
    print("Deployed Persistent Volume for " + openshift_project)

def deploy_persistent_volume_claim(persistent_volume_claim_api, yaml_file, openshift_project):
    print("Deploying Persistent Volume Claim for " + openshift_project)
    yaml_data = load_yaml(yaml_file)
    persistent_volume_claim_resp = persistent_volume_claim_api.create_namespaced_persistent_volume_claim(body=yaml_data, namespace=openshift_project)
    print(persistent_volume_claim_resp.metadata.self_link)
    print("Deployed Persistent Volume Claim for " + openshift_project)

def deploy_config_map(config_map_api, config_file, openshift_project):
    print("Deploying Config-Map for " + openshift_project)
    yaml_data = load_yaml(config_file)
    config_map_resp = config_map_api.create_namespaced_config_map(body=yaml_data, namespace=openshift_project)
    print(config_map_resp.metadata.self_link)
    print("Deployed Config-Map for " + openshift_project)

