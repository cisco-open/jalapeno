## Documentation describing link-state related virtual topology use cases and design
### Use cases include: internal traffic engineering, path steering, SRTE, explicit path TE, etc.

### LS Topology Model 
![ls topology](ls_topology.png)

The link-state topology processor (LS_Topology) creates a virtual representation of the network LSDB (the LS_Topology edge collection) by piecing together data from the existing Arango LSNode and LSLink collections.

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
  

