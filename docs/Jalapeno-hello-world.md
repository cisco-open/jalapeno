###  AQL queries to build a Hello World SR-App

#### The goal of this SR-App is to get the SR label stack for the least utilized path between Node 0007 and Node 0019:

#### Prior to setting up the disjoint query, we'll run a few other queries to gather data on the topology

1. Query to get the full link-state topology
```
for l in LSv4_Topology return l
```
![example topology](diagrams/example-topology.png)

2. Query to determine shortest-path hop-count from R01 to R05, and whether we have an ECMP pagths:
```
RETURN LENGTH(
FOR v IN OUTBOUND 
SHORTEST_PATH 'LSNodeDemo/0000.0000.0007' TO 'LSNodeDemo/0000.0000.0019' LSv4_Topology
 Return v
)
```
Output
```
[
  4
]
```
The AQL Output of 4 includes the source node.  From a routing perspective the source node is not includes, so our hop-count is 3

2. Query for ECMP paths should they exist.  The '3..3' notation gives us the hop count:
```
FOR v, e, p IN 3..3 OUTBOUND "LSNode/10.0.0.1" LS_Topology
     FILTER v._id == "LSNode/10.0.0.5"
       RETURN CONCAT_SEPARATOR(" -> ", p.vertices[*].RouterID)
```
Output
```
[
  "10.0.0.1 -> 10.0.0.8 -> 10.0.0.9 -> 10.0.0.5",
  "10.0.0.1 -> 10.0.0.3 -> 10.0.0.4 -> 10.0.0.5"
]
```

3. With that bit of background information we now Query for the shortest path which avoids R09
```
FOR v, e IN OUTBOUND 
SHORTEST_PATH 'LSNode/10.0.0.1' TO 'LSNode/10.0.0.5' LS_Topology
 FILTER v.RouterID != "10.0.0.9"
 Return [v.RouterID, e.RemotePrefixSID, e.AdjacencySID]
```
Output
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
The above query output can be used to create a disjoint SR LSP from R01 to R05, which avoids R09.  The 
SR-LSP may be executed via imposing label stack {100003, 100005}.  The output also includes SR Adjacency SID data should the user wish to force traffic out a specific link:


Just for fun, a query showing all paths from R01 to R05 that can be completed in 5 hops or less:
```
FOR v, e, p IN 1..6 OUTBOUND "LSNode/10.0.0.1" LS_Topology
     FILTER v._id == "LSNode/10.0.0.5"
       RETURN { "RouterID": p.vertices[*].RouterID, "PrefixSID": p.edges[*].RemotePrefixSID, "AdjSID": p.edges[*].AdjacencySID }
```



