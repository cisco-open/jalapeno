## Documentation describing link-state related virtual topology use cases and design
#### Use cases include: internal traffic engineering, path steering, SRTE, explicit path TE, etc.

#### LS Topology Diagram
![ls topology](ls_topology.png)

The link-state topology processor (LS_Topology) creates a virtual representation of the network LSDB and populates the LS_Topology edge collection by piecing together data from the existing Arango LSNode and LSLink collections.

Additionally, the [LS_Performance processor](ls_performance_processor.md) populates the LS_Topology edge collection with network performance and utilization statistics.

#### Sample Arango queries:

Entire link-state topology
```
for l in LSv4_Topology return l
```
Sample data:
```
  {
    "_key": "0000.0000.0000_10.1.1.0_0_0000.0000.0001_10.1.1.1_0",
    "_id": "LSv4_Topology/0000.0000.0000_10.1.1.0_0_0000.0000.0001_10.1.1.1_0",
    "_from": "LSNode/0000.0000.0000",
    "_to": "LSNode/0000.0000.0001",
    "_rev": "_bRJz1fG---",
    "timestamp": "Oct 16 19:50:07.000071",
    "local_interface_ip": [
      "10.1.1.0"
    ],
    "remote_interface_ip": [
      "10.1.1.1"
    ],
    "local_igp_id": "0000.0000.0000",
    "remote_igp_id": "0000.0000.0001",
    "local_node_asn": 100000,
    "remote_node_asn": 100000,
    "protocol": "IS-IS Level 2",
    "protocol_id": 2,
    "igp_metric": 1,
    "max_link_bw": 1000000000,
    "unresv_bw": [
      0,
      0,
      0,
      0,
      0,
      0,
      0,
      0,
      0,
      0,
      0,
      0,
      0,
      0,
      0,
      0
    ],
    "te_metric": 1,
    "srlg": null,
    "adjacency_sid": [
      {
        "flags": 112,
        "sid": 24012,
        "weight": 0
      },
      {
        "flags": 48,
        "sid": 24013,
        "weight": 0
      }
    ],
    "link_msd": [
      {
        "msd_type": 1,
        "msd_value": 10
      }
    ],
    "unidir_link_delay": 0,
    "unidir_link_delay_min_max": null,
    "unidir_delay_variation": 0,
    "unidir_packet_loss": 0,
    "unidir_residual_bw": 0,
    "unidir_available_bw": 0,
    "unidir_bw_utilization": 0,
    "app_resv_bw": "",
    "pq_resv_bw": "",
    "remote_msd": "[[{'msd_type': 1, 'msd_value': 10}]]",
    "link_delay": "",
    "local_msd": "[[{'msd_type': 1, 'msd_value': 10}]]",
    "remote_prefix_sid": 100001,
    "local_prefix_sid": 100000,
    "LocalPrefixInfo": [
      {
        "Prefix": "10.0.1.0",
        "Length": 32,
        "SID": 110000,
        "SRFlag": "N"
      },
      {
        "Prefix": "10.0.0.0",
        "Length": 32,
        "SID": 100000,
        "SRFlag": "N"
      }
    ],
    "remote_prefix_info": [
      {
        "Prefix": "10.0.0.1",
        "Length": 32,
        "SID": 100001,
        "SRFlag": "N"
      }
    ]
  },
```
Example shortest path query which returns Prefix SIDs for all nodes in the path:
```
FOR v IN OUTBOUND 
SHORTEST_PATH 'LSNode/10.0.0.6' TO 'LSNode/10.0.0.9' LS_Topology
 Return v.PrefixSID
```
Output:
```
[
  "100006",
  "100001",
  "100008",
  "100009"
]
```
Shortest path hop-count from one node to another:
```
RETURN LENGTH(
FOR v IN OUTBOUND 
SHORTEST_PATH 'LSNode/10.0.0.6' TO 'LSNode/10.0.0.9' LS_Topology
 Return v.PrefixSID
)
```
Output:
```
[
  4
]
```
Use the hop-count to create a query for ECMP paths (reference: https://www.arangodb.com/docs/stable/aql/examples-multiple-paths.html)
```
FOR v, e, p IN 3..3 OUTBOUND "LSNode/10.0.0.6" LS_Topology
     FILTER v._id == "LSNode/10.0.0.9"
       RETURN CONCAT_SEPARATOR(" -> ", p.vertices[*]._key)
```
Output:
```
[
  "10.0.0.6 -> 10.0.0.1 -> 10.0.0.8 -> 10.0.0.9",
  "10.0.0.6 -> 10.0.0.2 -> 10.0.0.8 -> 10.0.0.9"
]
```

