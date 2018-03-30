import sys, configparser
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
### Kafka Cluster Management
### Do they have a Kafka cluster or do we create one?
while True:
    kafka_exists = input("Do you have a Kafka cluster you would like to use (y/n): ")
    if kafka_exists in ('y', 'yes'):
        kafka_endpoint = input("Please enter the Kafka endpoint (i.e. 10.200.99.44:30902): ")
        config['VOLTRON']['kafka_endpoint'] = kafka_endpoint
        ### VALIDATE INPUT HERE
        ### SET THEIR KAFKA ENDPOINT IN NECESSARY FILES HERE
        break
    elif kafka_exists in ('n', 'no'):
        print("No worries! We're setting up a Kafka cluster for you in OpenShift now.")
        config['VOLTRON']['kafka_endpoint'] = host_ip + ":30902"
        ### SET UP KAFKA IN OPENSHIFT HERE
        break
    else:
        print("Please enter yes or no")

for i in range(89):
    cprint("-", 'blue', end=' ')
print("\n")
###########################################################################################################################

###########################################################################################################################
### InfluxDB Management
### Do they have an Influx instance or do we create one?
while True:
    influx_exists = input("Do you have a pre-existing InfluxDB instance you would like to use (y/n): ")
    if influx_exists in ('y', 'yes'):
        influx_endpoint = input("Please enter the InfluxDB endpoint (i.e. 10.200.99.44:30308): ")
        config['VOLTRON']['influx_endpoint'] = influx_endpoint
        ### VALIDATE INPUT HERE
        ### SET THEIR INFLUXDB ENDPOINT IN NECESSARY FILES HERE
        break
    elif influx_exists in ('n', 'no'):
        print("No worries! We're setting up a InfluxDB instance for you in OpenShift now.")
        config['VOLTRON']['influx_endpoint'] = host_ip + ":30308"
        ### SET UP INFLUXDB IN OPENSHIFT HERE
        break
    else:
        print("Please enter yes or no")

for i in range(89):
    cprint("-", 'blue', end=' ')
print("\n")
###########################################################################################################################

###########################################################################################################################
### OpenBMP Management
### Do they have a pre-existing OpenBMP setup?
while True:
    openbmp_exists = input("Do you have a pre-existing OpenBMP setup you would like to use (y/n): ")
    if openbmp_exists in ('y', 'yes'):
        openbmp_endpoint = input("Please enter the openbmp endpoint (i.e. 10.20.0.51:5000): ")
        config['VOLTRON']['openbmp_endpoint'] = openbmp_endpoint
        ### VALIDATE INPUT HERE
        ### SET THEIR OPENBMP ENDPOINT IN NECESSARY FILES HERE
        break
    elif openbmp_exists in ('n', 'no'):
        print("No worries! We'll configure your routers to send OpenBMP data, and we'll get that data into the Kafka cluster now.")
        config['VOLTRON']['openbmp_endpoint'] = host_ip + ":5000"
        ### SET UP OPENBMP HERE
        break
    else:
        print("Please enter yes or no")

for i in range(89):
    cprint("-", 'blue', end=' ')
print("\n")
###########################################################################################################################

###########################################################################################################################
### ArangoDB Management
### Do they have an ArangoDB instance or do we create one?
while True:
    arango_exists = input("Do you have a pre-existing instance of ArangoDB you would like to use (y/n): ")
    if arango_exists in ('y', 'yes'):
        arango_endpoint = input("Please enter the ArangoDB endpoint (i.e. 10.200.99.44:30852): ")
        config['VOLTRON']['arango_endpoint'] = arango_endpoint
        ### VALIDATE INPUT HERE
        ### SET THEIR ARANGO ENDPOINT IN NECESSARY FILES HERE
        break
    elif arango_exists in ('n', 'no'):
        print("No worries! We're setting up an ArangoDB instance for you in OpenShift now.")
        config['VOLTRON']['arango_endpoint'] = host_ip + ":30852"
        ### SET UP ARANGO IN OPENSHIFT HERE
        break
    else:
        print("Please enter yes or no")

for i in range(89):
    cprint("-", 'blue', end=' ')
print("\n")
###########################################################################################################################


with open('voltron.ini', 'w') as configfile:
    config.write(configfile)
