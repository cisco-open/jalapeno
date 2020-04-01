###  AQL queries to build an SDN Hello World App

#### The goal of this SR-App is to program a disjoint-path SR-LSP, which avoids R09:

![example topology](diagrams/example-topology.png)

1. Prior to setting up the disjoint query, we'll determine shortest-path hop-count from R01 to R05, and whether we have an ECMP pagths:
```
RETURN LENGTH(
FOR v IN OUTBOUND 
SHORTEST_PATH 'LSNode/10.0.0.1' TO 'LSNode/10.0.0.5' LS_Topology
 Return v
)
```
Output is 4-hops (including the source node):
```
[
  4
]
```
2. Query for ECMP paths should they exist.  In this case subtract 1 (see the '3..3' notation) from the hop-count result as source node is already accounted for:
```
FOR v, e, p IN 3..3 OUTBOUND "LSNode/10.0.0.1" LS_Topology
     FILTER v._id == "LSNode/10.0.0.5"
       RETURN CONCAT_SEPARATOR(" -> ", p.vertices[*]._key, p.vertices[*].PrefixSID)
```
Output
```
[
  "[\"10.0.0.1\",\"10.0.0.8\",\"10.0.0.9\",\"10.0.0.5\"] -> [\"100001\",\"100008\",\"100009\",\"100005\"]",
  "[\"10.0.0.1\",\"10.0.0.3\",\"10.0.0.4\",\"10.0.0.5\"] -> [\"100001\",\"100003\",\"100004\",\"100005\"]"
]
```
3. Query for the shortest path which avoids R09
```
FOR v, e IN OUTBOUND 
SHORTEST_PATH 'LSNode/10.0.0.1' TO 'LSNode/10.0.0.5' LS_Topology
 FILTER v.RouterID != "10.0.0.9"
 Return [v.RouterID, e.RemotePrefixSID, e.AdjacencySID]
```
The following output can be used to create an SR LSP which is disjoint from R09 via label stack 100003, 100005.  The output also includes SR Adjacency SID data should the user wish to force traffic out a specific link:
```
[
  [
    "10.0.0.1",
    null,
    null
  ],
  [
    "10.0.0.3",
    "100003",
    "BVL 0 24005, VL 0 24006"
  ],
  [
    "10.0.0.4",
    "100004",
    "BVL 0 24007, VL 0 24008"
  ],
  [
    "10.0.0.5",
    "100005",
    "BVL 0 24002, VL 0 24003"
  ]
]
```

Just for fun, a query showing all paths from R01 to R05 that can be completed in 5 hops or less:
```
FOR v, e, p IN 1..6 OUTBOUND "LSNode/10.0.0.1" LS_Topology
     FILTER v._id == "LSNode/10.0.0.5"
       RETURN { "RouterID": p.vertices[*].RouterID, "PrefixSID": p.edges[*].RemotePrefixSID, "AdjSID": p.edges[*].AdjacencySID }
```



