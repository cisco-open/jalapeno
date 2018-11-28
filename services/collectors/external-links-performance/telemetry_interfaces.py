"""This maps telemetry derived external interfaces with router IPs and interface IP addresses. In the future, this should be automated."""
telemetry_interface_mapper = {}
telemetry_interface_mapper['10.0.0.1_2.2.71.0'] = ['r1.622', 'GigabitEthernet0/0/0/1'] 
telemetry_interface_mapper['10.0.0.1_2.2.72.0'] = ['r1.622', 'GigabitEthernet0/0/0/2'] 
telemetry_interface_mapper['10.0.0.2_2.2.71.2'] = ['r2.622', 'GigabitEthernet0/0/0/1'] 
telemetry_interface_mapper['10.0.0.2_2.2.72.2'] = ['r2.622', 'GigabitEthernet0/0/0/2'] 

telemetry_interface_mapper['10.0.0.3_2.2.72.4'] = ['IAD_BR03.622', 'GigabitEthernet0/0/0/1'] 
telemetry_interface_mapper['10.0.0.3_2.2.78.0'] = ['IAD_BR03.622', 'GigabitEthernet0/0/0/2'] 
telemetry_interface_mapper['10.0.0.4_2.2.77.0'] = ['IAD_BR04.622', 'GigabitEthernet0/0/0/2'] 
telemetry_interface_mapper['10.0.0.4_2.2.78.2'] = ['IAD_BR04.622', 'GigabitEthernet0/0/0/1'] 


