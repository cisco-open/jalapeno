"""This maps telemetry derived internal interfaces with router IPs and interface IP addresses. In the future, this should be automated."""
telemetry_interface_mapper = {}
telemetry_interface_mapper['10.0.0.0_10.1.1.0'] = ['r0.622', 'GigabitEthernet0/0/0/3'] 
telemetry_interface_mapper['10.0.0.0_10.1.1.2'] = ['r0.622', 'GigabitEthernet0/0/0/4'] 
telemetry_interface_mapper['10.0.0.0_10.1.1.8'] = ['r0.622', 'GigabitEthernet0/0/0/5'] 
telemetry_interface_mapper['10.0.0.0_10.1.1.10'] = ['r0.622', 'GigabitEthernet0/0/0/6'] 

telemetry_interface_mapper['10.0.0.1_10.1.1.1'] = ['r1.622', 'GigabitEthernet0/0/0/0']
telemetry_interface_mapper['10.0.0.1_10.1.1.23'] = ['r1.622', 'GigabitEthernet0/0/0/4']
telemetry_interface_mapper['10.0.0.1_10.1.1.35'] = ['r1.622', 'GigabitEthernet0/0/0/3']

telemetry_interface_mapper['10.0.0.2_10.1.1.3'] = ['r2.622', 'GigabitEthernet0/0/0/0']
telemetry_interface_mapper['10.0.0.2_10.1.1.25'] = ['r2.622', 'GigabitEthernet0/0/0/4']
telemetry_interface_mapper['10.0.0.2_10.1.1.37'] = ['r2.622', 'GigabitEthernet0/0/0/3']
