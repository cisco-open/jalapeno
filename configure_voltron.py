import sys, configparser, os
from jinja2 import *
from colorama import init
init(strip=not sys.stdout.isatty()) # strip colors if stdout is redirected
from termcolor import cprint
from pyfiglet import figlet_format

def main():
    welcome_to_voltron()
    global config
    config = configparser.ConfigParser()
    config['VOLTRON'] = {}
    configured = check_configuration()
    if not configured: 
        request_configuration()
        with open('voltron.ini', 'w') as configfile:
            config.write(configfile)
    config.read('voltron.ini')
    configure_voltron()
    print("Voltron is configured! Deploy using `./deploy_voltron.sh`\n")

###########################################################################################################################
### Check if user has pre-configured endpoints
def check_configuration():
    while(True):
        configured_input = input("Has your voltron.ini file been filled out (y/n)? ")
        try:
            is_configured = process_input(configured_input)
        except ValueError:
            print("Please enter yes or no")
            continue
        break
    return is_configured
###########################################################################################################################

###########################################################################################################################
### Request endpoint IPs and more from user (set variables in voltron.ini)
def request_configuration():
    request_openshift()
    request_kafka_endpoint()
    request_arango_endpoint()
    request_influx_endpoint()
    request_network_asn()
###########################################################################################################################

###########################################################################################################################
### Set variables throughout Voltron (using Jinja templating and variables in voltron.ini)
def configure_voltron():
    configure_kafka_deployment()
    configure_arango_deployment()
    configure_openbmp_deployment()
    configure_telemetry_deployment()
    configure_topology_service()
    configure_performance_services()
###########################################################################################################################

###########################################################################################################################
### Host Management
def request_openshift():
    while True:
        host_ip = input("Please enter the host IP address where OpenShift will be deployed (i.e. 10.0.250.2): ")
        ### VALIDATE INPUT HERE
        if(host_ip != '10.0.250.2'):
            print("Please enter a valid IP address!")
        else:
            config['VOLTRON']['host_ip'] = host_ip
            break
    pretty_print_split()
###########################################################################################################################

###########################################################################################################################
### Kafka Cluster Management
### Do they have a Kafka cluster or do we create one?
def request_kafka_endpoint():
    while True:
        kafka_input = input("Do you have a Kafka cluster you would like to use (y/n): ")
        try:
            kafka_exists = process_input(kafka_input)
        except ValueError:
            print("Please enter yes or no")
            continue
        if kafka_exists:
            kafka_endpoint = input("Please enter the Kafka endpoint (i.e. 10.200.99.3:30902): ")
            config['VOLTRON']['kafka_endpoint'] = kafka_endpoint
            break
        elif not kafka_exists:
            print("No worries! We'll configure a Kafka cluster to be deployed in OpenShift now.")
            kafka_endpoint = config['VOLTRON']['host_ip'] + ":30902"
            config['VOLTRON']['kafka_endpoint'] = kafka_endpoint
            break
    pretty_print_split()
###########################################################################################################################

###########################################################################################################################
### ArangoDB Management
### Do they have an ArangoDB instance or do we create one?
def request_arango_endpoint():
    while True:
        arango_input = input("Do you have a pre-existing instance of ArangoDB you would like to use (y/n): ")
        try:
            arango_exists = process_input(arango_input)
        except ValueError:
            print("Please enter yes or no")
            continue
        if arango_exists:
            arango_endpoint = input("Please enter the ArangoDB endpoint (i.e. 10.200.99.3:30852): ")
            config['VOLTRON']['arango_endpoint'] = arango_endpoint
            break
        elif not arango_exists:
            print("No worries! We'll configure an ArangoDB instance to be deployed in OpenShift now.")
            arango_endpoint = config['VOLTRON']['host_ip'] + ":30852"
            config['VOLTRON']['arango_endpoint'] = arango_endpoint
            break
    pretty_print_split()
###########################################################################################################################

###########################################################################################################################
### InfluxDB Management
### Do they have an Influx instance or do we create one?
def request_influx_endpoint():
    while True:
        influx_input = input("Do you have a pre-existing InfluxDB instance you would like to use (y/n): ")
        try:
            influx_exists = process_input(influx_input)
        except ValueError:
            print("Please enter yes or no")
            continue
        if influx_exists:
            influx_endpoint = input("Please enter the InfluxDB endpoint (i.e. 10.200.99.3:30308): ")
            config['VOLTRON']['influx_endpoint'] = influx_endpoint
            break
        elif not influx_exists:
            print("No worries! We'll configure an InfluxDB instance to be deployed in OpenShift now.")
            influx_endpoint = config['VOLTRON']['host_ip'] + ":30308"
            config['VOLTRON']['influx_endpoint'] = influx_endpoint
            break
    pretty_print_split()
###########################################################################################################################

###########################################################################################################################
### ASN Management
def request_network_asn():
    asn_input = input("What is(are) your internal network ASN(s)? ")
    print("Thanks! We'll configure services to recognize inputted ASN(s).")
    config['VOLTRON']['network_asn'] = asn_input
    pretty_print_split()
###########################################################################################################################

###########################################################################################################################
### ArangoDB automation
### Rendering Arango's infrastructure YAML file with host_IP
def configure_arango_deployment():
    print("Configuring Arango deployment")
    templateLoader = FileSystemLoader(searchpath="./infra/templates/arangodb/")
    templateEnv = Environment(loader=templateLoader)
    TEMPLATE_FILE = "arangodb_apps_pv_template.yaml"
    template = templateEnv.get_template(TEMPLATE_FILE)
    context = {
        'host_ip': config['VOLTRON']['host_ip'],
    }
    outputText = template.render(context)
    dirname = os.path.dirname(os.path.abspath(__file__))
    arango_apps_persistent_volume_yaml = os.path.join(dirname, 'infra', 'arangodb', 'arangodb_apps_pv.yaml')
    with open(arango_apps_persistent_volume_yaml, "w") as file_handler:
        file_handler.write(outputText)

    TEMPLATE_FILE = "arangodb_pv_template.yaml"
    template = templateEnv.get_template(TEMPLATE_FILE)
    context = {
        'host_ip': config['VOLTRON']['host_ip'],
    }
    outputText = template.render(context)
    dirname = os.path.dirname(os.path.abspath(__file__))
    arango_persistent_volume_yaml = os.path.join(dirname, 'infra', 'arangodb', 'arangodb_pv.yaml')
    with open(arango_persistent_volume_yaml, "w") as file_handler:
        file_handler.write(outputText)
###########################################################################################################################

###########################################################################################################################
### Kafka automation
### Rendering Kafka's/Zookeeper's infrastructure YAML files with host_IP
def configure_kafka_deployment():
    print("Configuring Kafka cluster deployment")
    context = {
        'host_ip': config['VOLTRON']['host_ip'],
        'kafka_endpoint': config['VOLTRON']['kafka_endpoint'],
    }
    templateLoader = FileSystemLoader(searchpath="./infra/templates/kafka/")
    templateEnv = Environment(loader=templateLoader)
    dirname = os.path.dirname(os.path.abspath(__file__))

    TEMPLATE_FILE = "zookeeper_pv_template.yaml"
    template = templateEnv.get_template(TEMPLATE_FILE)
    outputText = template.render(context)
    zookeeper_persistent_volume_yaml = os.path.join(dirname, 'infra', 'kafka', 'zookeeper_pv.yaml')
    with open(zookeeper_persistent_volume_yaml, "w") as file_handler:
        file_handler.write(outputText)

    TEMPLATE_FILE = "kafka_pv_template.yaml"
    template = templateEnv.get_template(TEMPLATE_FILE)
    outputText = template.render(context)
    kafka_persistent_volume_yaml = os.path.join(dirname, 'infra', 'kafka', 'kafka_pv.yaml')
    with open(kafka_persistent_volume_yaml, "w") as file_handler:
        file_handler.write(outputText)

    TEMPLATE_FILE = "kafka_ss_template.yaml"
    template = templateEnv.get_template(TEMPLATE_FILE)
    outputText = template.render(context)
    kafka_stateful_set_yaml = os.path.join(dirname, 'infra', 'kafka', 'kafka_ss.yaml')
    with open(kafka_stateful_set_yaml, "w") as file_handler:
        file_handler.write(outputText)
###########################################################################################################################

###########################################################################################################################
### OpenBMP automation
### Rendering OpenBMP's config and service files with v0-vm0 IP and Kafka endpoint
def configure_openbmp_deployment():
    print("Configuring OpenBMP deployment")
    context = {
        'kafka_endpoint': config['VOLTRON']['kafka_endpoint'],
    }
    templateLoader = FileSystemLoader(searchpath="./infra/templates/openbmpd/")
    templateEnv = Environment(loader=templateLoader)
    dirname = os.path.dirname(os.path.abspath(__file__))

    TEMPLATE_FILE = "deploy_openbmp.py"
    template = templateEnv.get_template(TEMPLATE_FILE)
    outputText = template.render(context)
    openbmpd_service_template = os.path.join(dirname, 'infra', 'openbmpd', 'deploy_openbmp.py')
    with open(openbmpd_service_template, "w") as file_handler:
        file_handler.write(outputText)
###########################################################################################################################

###########################################################################################################################
### Telemetry automation
### Rendering Pipeline infrastructure config file with kafka_endpoint
def configure_telemetry_deployment():
    print("Configuring Telemetry deployment")
    templateLoader = FileSystemLoader(searchpath="./infra/templates/telemetry/")
    templateEnv = Environment(loader=templateLoader)
    TEMPLATE_FILE = "pipeline_template.conf"
    template = templateEnv.get_template(TEMPLATE_FILE)
    context = {
        'kafka_endpoint': config['VOLTRON']['kafka_endpoint'],
    }
    outputText = template.render(context)
    dirname = os.path.dirname(os.path.abspath(__file__))
    pipeline_config = os.path.join(dirname, 'infra', 'telemetry', 'pipeline', 'pipeline.conf')
    with open(pipeline_config, "w") as file_handler:
        file_handler.write(outputText)
###########################################################################################################################

###########################################################################################################################
### Topology automation
### Rendering Topology Collector Service YAML file with host_IP and port
def configure_topology_service():
    print("Configuring Topology Collector service")
    templateLoader = FileSystemLoader(searchpath="./services/templates/collectors/")
    templateEnv = Environment(loader=templateLoader)
    TEMPLATE_FILE = "topology_dp_template.yaml"
    template = templateEnv.get_template(TEMPLATE_FILE)
    context = {
        'kafka_endpoint': config['VOLTRON']['kafka_endpoint'],
        'network_asn' : config['VOLTRON']['network_asn'],
    }
    outputText = template.render(context)
    dirname = os.path.dirname(os.path.abspath(__file__))
    topology_deployment = os.path.join(dirname, 'services', 'collectors', 'topology', 'topology_dp.yaml')
    with open(topology_deployment, "w") as file_handler:
        file_handler.write(outputText)
###########################################################################################################################

###########################################################################################################################
### Performance Collection Services
### Rendering Performance Collection Services YAML files with host_IP and port
def configure_performance_services():
    print("Configuring Performance Collection services")
    templateLoader = FileSystemLoader(searchpath="./services/templates/configs/")
    templateEnv = Environment(loader=templateLoader)
    TEMPLATE_FILE = "arangoconfig.yaml"
    template = templateEnv.get_template(TEMPLATE_FILE)
    context = {
        'arango_endpoint': config['VOLTRON']['arango_endpoint'],
    }
    outputText = template.render(context)

    dirname = os.path.dirname(os.path.abspath(__file__))
    internal_links_arango = os.path.join(dirname, 'services', 'collectors', 'internal-links-performance', 'configs', 'arangoconfig.py')
    with open(internal_links_arango, "w") as file_handler:
        file_handler.write(outputText)

    dirname = os.path.dirname(os.path.abspath(__file__))
    external_links_arango = os.path.join(dirname, 'services', 'collectors', 'external-links-performance', 'configs', 'arangoconfig.py')
    with open(external_links_arango, "w") as file_handler:
        file_handler.write(outputText)

    dirname = os.path.dirname(os.path.abspath(__file__))
    epe_edges_arango = os.path.join(dirname, 'services', 'collectors', 'epe-edges', 'configs', 'arangoconfig.py')
    with open(epe_edges_arango, "w") as file_handler:
        file_handler.write(outputText)

    dirname = os.path.dirname(os.path.abspath(__file__))
    epe_paths_arango = os.path.join(dirname, 'services', 'collectors', 'epe-paths', 'configs', 'arangoconfig.py')
    with open(epe_paths_arango, "w") as file_handler:
        file_handler.write(outputText)


    TEMPLATE_FILE = "influxconfig.yaml"
    template = templateEnv.get_template(TEMPLATE_FILE)
    influx_endpoint = config['VOLTRON']['influx_endpoint']
    influx_url = influx_endpoint.split(':')
    context = {
        'influx_ip': influx_url[0],
        'influx_port': influx_url[1],
    }
    outputText = template.render(context)

    dirname = os.path.dirname(os.path.abspath(__file__))
    internal_links_influx = os.path.join(dirname, 'services', 'collectors', 'internal-links-performance', 'configs', 'influxconfig.py')
    with open(internal_links_influx, "w") as file_handler:
        file_handler.write(outputText)

    dirname = os.path.dirname(os.path.abspath(__file__))
    external_links_influx = os.path.join(dirname, 'services', 'collectors', 'external-links-performance', 'configs', 'influxconfig.py')
    with open(external_links_influx, "w") as file_handler:
        file_handler.write(outputText)
###########################################################################################################################

###########################################################################################################################
### ASCII Art for Fun
def welcome_to_voltron():
    for i in range(534):
        cprint("-", 'blue', end=' ')

    print("\n")
    cprint(figlet_format('Voltron', font='starwars'),
           'white', attrs=['bold'])
    for i in range(534):
        cprint("-", 'blue', end=' ')

    print("\n")
    print("Welcome to Voltron")
    for i in range(178):
        cprint("-", 'blue', end=' ')

def pretty_print_split():
    for i in range(89):
        cprint("-", 'blue', end=' ')
    print("\n")
###########################################################################################################################

###########################################################################################################################
### Input Processing
def process_input(input):
    if input in ('y', 'yes', 'Y', 'Yes'):
        return True
    elif input in ( 'n', 'no', 'N', 'No'):
        return False
    else:
        raise ValueError('User entered invalid input.')
###########################################################################################################################

if __name__ == '__main__':
    main()