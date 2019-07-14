"""This maps telemetry derived internal interfaces with router IPs and interface IP addresses. In the future, this should be automated."""
telemetry_interface_mapper = {}
telemetry_interface_mapper['10.0.0.0_10.1.1.0'] = ['ORD_TR00.6.3.3', 'GigabitEthernet0/0/0/3'] 
telemetry_interface_mapper['10.0.0.0_10.1.1.2'] = ['ORD_TR00.6.3.3', 'GigabitEthernet0/0/0/4'] 

telemetry_interface_mapper['10.0.0.1_10.1.1.1'] = ['ORD_PR01', 'GigabitEthernet0/0/0/0']
telemetry_interface_mapper['10.0.0.1_10.1.1.5'] = ['ORD_PR01', 'GigabitEthernet0/0/0/5']

telemetry_interface_mapper['10.0.0.2_10.1.1.3'] = ['ORD_PR02', 'GigabitEthernet0/0/0/0']
telemetry_interface_mapper['10.0.0.2_10.1.1.7'] = ['ORD_PR02', 'GigabitEthernet0/0/0/5']
