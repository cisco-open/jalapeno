# Latency Service

Latency Calculation Process 
(done for each path in the Paths collection)
- get the label stack for a path 
- create an MPLS packet with the label stack incorporated
- start listening on our egress interface
- send MPLS packet out
- record request/reply time
- calculate latency
- upsert the latency into the path document in the collection

## Details
A "Paths" collection exists in our ArangoDB instance. The collection is built by the "Paths" service.

### Label Stack Generation
First, we need to get the label stack for each path in the Paths collection. We will need these 
label stacks to create a forceful ping over a specific path.

The "label_generator" script runs the necessary AQL query to generate the label stacks. It generates a stack
for each path from our source (specified in configs/queryconfig.py) to each of the destinations (specified in configs/prefixes.txt).

### MPLS Packet Creation
The "network_latency" script takes the label stacks and creates an MPLS packet for each one (using the "MPLS_packet_manager" script). It then calls on the "network_sniff" script to begin listening on the egress interface, and spawns a new process to send the MPLS packet out.

### Latency Calculation and Upsert
As the "network_latency" script sends out each MPLS packet, the "network_sniff" script records request and reply times. The difference is the latency for the given path. The "network_sniff" script then upserts that latency into the Paths collection for the specific path. 
