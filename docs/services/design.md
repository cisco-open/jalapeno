# vService + Topology interactions
![Design Decisions]("./methods.png")

The picture above depicts two possible methods for vService interaction with the topology.
- The first involves vServices updating values on graph edges directly.
- The second would have each vService create it's own graph based off the topology with the data it has learned so far.

We have decided to go with the first method for the following reasons:
-  Say you want to use Latency and Utilization to find the path, then one would have to do three queries: the latency graph, the utilization graph and the topology graph.
- Pruning edges - A latency decision might use a specific link. If it goes down, it will be updated in the topology database, but would not be necessarily reflected right away in the vService graph.
- There will be _A LOT_ of edges. There are a LOT of prefixes...
- A front end query-like vService would have to know how to interpret each backend vService's custom database. Combining these results will also be hard. What if you only get partial information from one, and more complete information from another etc.
- Translating from the subgraph to a label stack would be hard. The optimal path in the view of the subgraph might include "rtrp2 and AS6000"... which would make the corresponding query awkward/hard.
