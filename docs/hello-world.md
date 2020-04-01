###  An SDN Hello World App

![example topology](diagrams/example-topology.png)

Sample Arango queries to assemble the data needed to program a disjoint-path LSP:

1. Determine shortest-path hop-count from R01 to R05:
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
2. Optional query for ECMP paths should they exist.  In this case subtract 1 (see the '3..3' notation) from the hop-count result as source node is already accounted for:
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
3. Query for shortest path which avoids R09
```
query
```
Output
```
output
```


