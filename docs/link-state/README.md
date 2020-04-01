## Documentation describing link-state related virtual topology use cases and design
#### Use cases include: internal traffic engineering, path steering, SRTE, explicit path TE, etc.

#### LS Topology Diagram
![ls topology](ls_topology.png)

The link-state topology processor (LS_Topology) creates a virtual representation of the network LSDB and populates the LS_Topology edge collection by piecing together data from the existing Arango LSNode and LSLink collections.

Additionally, the [LS_Performance processor](ls_performance_processor.md) populates the LS_Topology edge collection with network performance and utilization statistics.

#### Sample Arango queries:

Entire link-state topology
```
for l in LS_Topology return l
```
Sample data:
```
  {
    "_key": "10.0.0.10_10.1.1.29_10.1.1.28_10.0.0.1",
    "_id": "LS_Topology/10.0.0.10_10.1.1.29_10.1.1.28_10.0.0.1",
    "_from": "LSNode/10.0.0.10",
    "_to": "LSNode/10.0.0.1",
    "_rev": "_aRYG1w---_",
    "LocalRouterID": "10.0.0.10",
    "RemoteRouterID": "10.0.0.1",
    "Protocol": "IS-IS_L2",
    "IGPID": "0000.0000.0010.0000",
    "ASN": "100000",
    "FromInterfaceIP": "10.1.1.29",
    "ToInterfaceIP": "10.1.1.28",
    "IGPMetric": "1",
    "TEMetric": "1",
    "AdminGroup": "0",
    "MaxLinkBW": "1000000",
    "MaxResvBW": "0",
    "UnResvBW": "0, 0, 0, 0, 0, 0, 0, 0",
    "AdjacencySID": "VL 0 24001",
    "AppResvBW": "",
    "PQResvBW": "",
    "LocalMaxSIDDepth": "",
    "Out-Octets": "",
    "In-Discards": "",
    "Out-Discards": "",
    "RemoteMaxSIDDepth": "",
    "In-Octets": "",
    "Link-Delay": "",
    "RemotePrefixSID": "100001",
    "LocalPrefixSID": "100010",
    "SRLG": "",
    "Adjacencies": [
      {
        "adjacency_sid": "24001",
        "flags": "VL",
        "weight": "0"
      }
    ],
    "Percent_Util_Outbound": 0,
    "Percent_Util_Inbound": 0,
    "In_Multicast_Pkts": 0,
    "In_Broadcast_Pkts": 0,
    "In_Errors": 0,
    "In_Discards": 0,
    "Speed": 1,
    "In_Octets": 0,
    "Out_Discards": 0,
    "Out_Unicast_Pkts": 0,
    "Out_Broadcast_Pkts": 0,
    "In_Unicast_Pkts": 0,
    "Out_Multicast_Pkts": 0,
    "Out_Errors": 0,
    "Out_Octets": 0
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

