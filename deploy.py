import sys, configparser, os
from jinja2 import *
from colorama import init
init(strip=not sys.stdout.isatty()) # strip colors if stdout is redirected
from termcolor import cprint
from pyfiglet import figlet_format


###########################################################################################################################
### ASCII Art for Fun
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
###########################################################################################################################


config = configparser.ConfigParser()
config['VOLTRON'] = {}

###########################################################################################################################
### Host Management
while True:
    host_ip = input("Please enter the host IP address where OpenShift will be deployed (i.e. 10.200.99.44): ")
    ### VALIDATE INPUT HERE
    if(host_ip != '10.200.99.44'):
        print("Please enter a valid IP address!")
    else:
        config['VOLTRON']['host_ip'] = host_ip
        break

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


###########################################################################################################################
### Kafka Cluster Management
### Do they have a Kafka cluster or do we create one?
while True:
    kafka_input = input("Do you have a Kafka cluster you would like to use (y/n): ")
    try:
        kafka_exists = process_input(kafka_input)
    except ValueError:
        print("Please enter yes or no")
        continue
    if kafka_exists:
        kafka_endpoint = input("Please enter the Kafka endpoint (i.e. 10.200.99.44:30902): ")
        config['VOLTRON']['kafka_endpoint'] = kafka_endpoint
        ### VALIDATE INPUT HERE
        ### SET THEIR KAFKA ENDPOINT IN NECESSARY FILES HERE
        break
    elif not kafka_exists:
        print("No worries! We're setting up a Kafka cluster for you in OpenShift now.")
        kafka_endpoint = host_ip + ":30902"
        config['VOLTRON']['kafka_endpoint'] = kafka_endpoint
        ### SET UP KAFKA IN OPENSHIFT HERE
        break

for i in range(89):
    cprint("-", 'blue', end=' ')
print("\n")
###########################################################################################################################

###########################################################################################################################
### InfluxDB Management
### Do they have an Influx instance or do we create one?
while True:
    influx_input = input("Do you have a pre-existing InfluxDB instance you would like to use (y/n): ")
    try:
        influx_exists = process_input(influx_input)
    except ValueError:
        print("Please enter yes or no")
        continue
    if influx_exists:
        influx_endpoint = input("Please enter the InfluxDB endpoint (i.e. 10.200.99.44:30308): ")
        config['VOLTRON']['influx_endpoint'] = influx_endpoint
        ### VALIDATE INPUT HERE
        ### SET THEIR INFLUXDB ENDPOINT IN NECESSARY FILES HERE
        break
    elif not influx_exists:
        print("No worries! We're setting up a InfluxDB instance for you in OpenShift now.")
        influx_endpoint = host_ip + ":30308"
        config['VOLTRON']['influx_endpoint'] = influx_endpoint
        ### SET UP INFLUXDB IN OPENSHIFT HERE
        break

for i in range(89):
    cprint("-", 'blue', end=' ')
print("\n")
###########################################################################################################################

###########################################################################################################################
### OpenBMP Management
### Do they have a pre-existing OpenBMP setup?
while True:
    openbmp_input = input("Do you have a pre-existing OpenBMP setup you would like to use (y/n): ")
    try:
        openbmp_exists = process_input(openbmp_input)
    except ValueError:
        print("Please enter yes or no")
        continue
    if openbmp_exists:
        openbmp_endpoint = input("Please enter the openbmp endpoint (i.e. 10.20.0.51:5000): ")
        config['VOLTRON']['openbmp_endpoint'] = openbmp_endpoint
        ### VALIDATE INPUT HERE
        ### SET THEIR OPENBMP ENDPOINT IN NECESSARY FILES HERE
        break
    elif not openbmp_exists:
        print("No worries! We'll configure your routers to send OpenBMP data, and we'll get that data into the Kafka cluster now.")
        openbmp_endpoint = host_ip + ":5000"
        config['VOLTRON']['openbmp_endpoint'] = openbmp_endpoint
        ### SET UP OPENBMP HERE
        break

for i in range(89):
    cprint("-", 'blue', end=' ')
print("\n")
###########################################################################################################################

###########################################################################################################################
### ArangoDB Management
### Do they have an ArangoDB instance or do we create one?
while True:
    arango_input = input("Do you have a pre-existing instance of ArangoDB you would like to use (y/n): ")
    try:
        arango_exists = process_input(arango_input)
    except ValueError:
        print("Please enter yes or no")
        continue
    if arango_exists:
        arango_endpoint = input("Please enter the ArangoDB endpoint (i.e. 10.200.99.44:30852): ")
        config['VOLTRON']['arango_endpoint'] = arango_endpoint
        ### VALIDATE INPUT HERE
        ### SET THEIR ARANGO ENDPOINT IN NECESSARY FILES HERE
        break
    elif not arango_exists:
        print("No worries! We're setting up an ArangoDB instance for you in OpenShift now.")
        arango_endpoint = host_ip + ":30852"
        config['VOLTRON']['arango_endpoint'] = arango_endpoint
        ### SET UP ARANGO IN OPENSHIFT HERE
        break

for i in range(89):
    cprint("-", 'blue', end=' ')
print("\n")
###########################################################################################################################


###########################################################################################################################
with open('voltron.ini', 'w') as configfile:
    config.write(configfile)
###########################################################################################################################


###########################################################################################################################
### ArangoDB automation
### Rendering Arango's infrastructure YAML file with host_IP
templateLoader = FileSystemLoader(searchpath="./templates/infra/arangodb/")
templateEnv = Environment(loader=templateLoader)
TEMPLATE_FILE = "arangodb_apps_pv_template.yaml"
template = templateEnv.get_template(TEMPLATE_FILE)
context = {
    'host_ip': host_ip,
}
outputText = template.render(context)
dirname = os.path.dirname(os.path.abspath(__file__))
arango_apps_persistent_volume_yaml = os.path.join(dirname, 'infra', 'arangodb', 'arangodb_apps_pv.yaml')
with open(arango_apps_persistent_volume_yaml, "w") as file_handler:
    file_handler.write(outputText)

TEMPLATE_FILE = "arangodb_pv_template.yaml"
template = templateEnv.get_template(TEMPLATE_FILE)
context = {
    'host_ip': host_ip,
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
context = {
    'host_ip': host_ip,
}
templateLoader = FileSystemLoader(searchpath="./templates/infra/kafka/")
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

###########################################################################################################################


###########################################################################################################################
### OpenBMP automation
### Rendering OpenBMP's config and service files with v0-vm0 IP and Kafka endpoint
context = {
    'kafka_endpoint': kafka_endpoint,
}
templateLoader = FileSystemLoader(searchpath="./templates/infra/openbmpd/")
templateEnv = Environment(loader=templateLoader)
dirname = os.path.dirname(os.path.abspath(__file__))

TEMPLATE_FILE = "openbmpd_service_template"
template = templateEnv.get_template(TEMPLATE_FILE)
outputText = template.render(context)
openbmpd_service_template = os.path.join(dirname, 'infra', 'openbmpd', 'openbmpd.service')
with open(openbmpd_service_template, "w") as file_handler:
    file_handler.write(outputText)

TEMPLATE_FILE = "openbmpd_process_vars_template"
template = templateEnv.get_template(TEMPLATE_FILE)
outputText = template.render(context)
openbmpd_process_vars = os.path.join(dirname, 'infra', 'openbmpd', 'openbmpd')
with open(openbmpd_process_vars, "w") as file_handler:
    file_handler.write(outputText)

###########################################################################################################################

###########################################################################################################################
### Telemetry automation
### Rendering Pipeline infrastructure config file with kafka_endpoint
templateLoader = FileSystemLoader(searchpath="./templates/infra/telemetry/")
templateEnv = Environment(loader=templateLoader)
TEMPLATE_FILE = "pipeline_template.conf"
template = templateEnv.get_template(TEMPLATE_FILE)
context = {
    'kafka_endpoint': kafka_endpoint,
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
templateLoader = FileSystemLoader(searchpath="./templates/services/collectors/topology/")
templateEnv = Environment(loader=templateLoader)
TEMPLATE_FILE = "topology_dp_template.yaml"
template = templateEnv.get_template(TEMPLATE_FILE)
context = {
    'kafka_endpoint': kafka_endpoint,
}
outputText = template.render(context)
dirname = os.path.dirname(os.path.abspath(__file__))
pipeline_config = os.path.join(dirname, 'services', 'collectors', 'topology', 'topology_dp.yaml')
with open(pipeline_config, "w") as file_handler:
    file_handler.write(outputText)
###########################################################################################################################




####
# call voltron/infra/deploy_infrastructure.py
# call voltron/services/deploy_service.py
# call voltron/client/deploy_client.py
####
