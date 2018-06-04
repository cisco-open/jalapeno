"""This maps telemetry derived interfaces with Arango LinkEdgeV4 keys.
This enables upsertion of bandwidth values into corresponding link edges.
In the future, this should be automated."""

interface_linkedge_mapper = {}
interface_linkedge_mapper['r0.622_GigabitEthernet0/0/0/4'] = '10.0.0.32_10.0.0.33'
interface_linkedge_mapper['r0.622_GigabitEthernet0/0/0/5'] = '10.0.0.34_10.0.0.35'

interface_linkedge_mapper['r1.622_GigabitEthernet0/0/0/1'] = '10.0.0.36_10.0.0.37'
interface_linkedge_mapper['r1.622_GigabitEthernet0/0/0/2'] = '10.0.0.38_10.0.0.39'

interface_linkedge_mapper['r2.622_GigabitEthernet0/0/0/1'] = '10.0.0.40_10.0.0.41'
interface_linkedge_mapper['r2.622_GigabitEthernet0/0/0/2'] = '10.0.0.42_10.0.0.43'
