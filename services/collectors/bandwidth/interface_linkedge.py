"""This maps telemetry derived interfaces with Arango LinkEdgeV4 keys.
This enables upsertion of bandwidth values into corresponding link edges.
In the future, this should be automated."""

interface_linkedge_mapper = {}
interface_linkedge_mapper['v0-t0_GigabitEthernet0/0/0/1'] = '2.2.2.2_2.2.2.3'
interface_linkedge_mapper['v0-t0_GigabitEthernet0/0/0/2'] = '2.2.2.4_2.2.2.5'

interface_linkedge_mapper['v0-l1_GigabitEthernet0/0/0/1'] = '2.2.2.6_2.2.2.7'
interface_linkedge_mapper['v0-l1_GigabitEthernet0/0/0/2'] = '2.2.2.8_2.2.2.9'

interface_linkedge_mapper['v0-l2_GigabitEthernet0/0/0/1'] = '2.2.2.10_2.2.2.11'
interface_linkedge_mapper['v0-l2_GigabitEthernet0/0/0/2'] = '2.2.2.12_2.2.2.13'

interface_linkedge_mapper['v0-p3_GigabitEthernet0/0/0/2'] = '2.2.2.14_2.2.2.15'
interface_linkedge_mapper['v0-p3_GigabitEthernet0/0/0/3'] = '2.2.2.16_2.2.2.17'
interface_linkedge_mapper['v0-p3_GigabitEthernet0/0/0/4'] = '2.2.2.18_2.2.2.19'

interface_linkedge_mapper['v0-p4_GigabitEthernet0/0/0/2'] = '2.2.2.20_2.2.2.21'
interface_linkedge_mapper['v0-p4_GigabitEthernet0/0/0/3'] = '2.2.2.22_2.2.2.23'
interface_linkedge_mapper['v0-p4_GigabitEthernet0/0/0/4'] = '2.2.2.24_2.2.2.25'
