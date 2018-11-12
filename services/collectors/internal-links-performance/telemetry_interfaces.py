"""This maps telemetry derived interfaces with router IPs and interface IP addresses. In the future, this should be automated."""

telemetry_interface_mapper = {}

telemetry_interface_mapper['10.0.0.0_10.1.1.0'] = ['r0.622', 'GigabitEthernet0/0/0/3'] 
telemetry_interface_mapper['10.0.0.0_10.1.1.2'] = ['r0.622', 'GigabitEthernet0/0/0/4'] 
telemetry_interface_mapper['10.0.0.0_10.1.1.8'] = ['r0.622', 'GigabitEthernet0/0/0/5'] 
telemetry_interface_mapper['10.0.0.0_10.1.1.10'] = ['r0.622', 'GigabitEthernet0/0/0/6'] 

