# Telemetry Automation

This folder contains scripts to automate the enablement and operation of telemetry against host devices.  
Can run from any host which has network connectivity to the configured devices.  
Relies on files from [infra/telemetry](../infra/telemetry)

## Support
* IOS-XR

## Installation
```bash
# if pipenv is not installed...
pip install pipenv
# endif
pipenv --three install
```

## Configuration
```hosts.json``` is the base configuration file which these toolings operate against. [hosts.json.example](telemetry/hosts.json.example) contains the configuration structure which is expected.

```bash
cp hosts.json.example hosts.json
```

### Host
```netmiko_network``` and ```netmiko_linux``` are the two main configuration sections per host device that will be configured. This tooling relies heavily on [netmiko](https://github.com/ktbyers/netmiko) for its functionalities. Configuration items correspond 1:1 with a netmiko configuration for ease of use, and because netmiko inherently requires configuration for automation of devices and thus is a fine configuration template.

#### netmiko_network
The netmiko configuration which corresponds to the device that is being configured. For instance, for IOS-XR, ```cisco_xr``` is the ```device_type```, and for NX, ```cisco_nx``` is the ```device_type```. Anything utilizing these credentials needs access to the data plane and control plane of the network device itself.

#### netmiko_linux
This configuration section essentially implies what we refer to as "guestshell" on the network devices today. It should, ideally, utilize an unprivileged user. This is where we can operationalize Linux-based agents, toolings, and binaries and operate as if we were on a regular host system. This makes it very easy to develop against, and enables useful functionality like the ability to run Pipeline on the device itself (which we do!).

## configure_telemetry.py

Configures MDT on to IOS-XR network devices. Utilizes [infra/telemetry/config_xr](../infra/telemetry/config_xr) as configuration file. Configures every device in the configuration file, does not filter yet.

### Usage
```bash
pipenv shell
python configure_telemetry.py
```

## enable_guestshell.py

Enables the Linux "guestshell" on IOS-XR network devices. Checks if the guestshell SSH port is open and guestshell credentials are valid. If the guestshell SSH port is not open then it configures guestshell and creates an unprivileged user in guestshell based upon the one specified in the ```netmiko_linux``` (guestshell) configuration section. This user does not have privileges to configure the network device. Checks against all configured devices.

### Usage
```bash
pipenv shell
python enable_guestshell.py
```

## manage_pipeline.py

Provides functionality to manage Pipeline instances on Linux hosts. Utilizes the [Pipeline configuration](../infra/telemetry/pipeline) under ```/infra/```.  
Uses the ```netmiko_linux``` configuration per device.

* Provision
* Remove
* Reset (Remove/Provision)
* Start
* Stop
* Restart (Stop/Start)

### Usage
```bash
pipenv shell
# usage: manage_pipeline.py [-h] [--hostnames HOSTNAMES [HOSTNAMES ...]] action
# Operate against all configured hosts
python manage_pipeline.py provision
# Operate against a subset of configured hosts
python manage_pipeline.py --hostnames 127.0.0.1 mydevice provision
```
