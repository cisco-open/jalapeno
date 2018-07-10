# Bandwidth Service

Bandwidth Calculation Process 
(done for each path in the Paths collection)
- collect telemetry data
- aggregate necessary fields from telemetry in some sort of a structure
- execute bandwidth calculation for rolling average (specifically bytes sent/received) 
- upsert the bandwidth into the path document in the collection

## Details
A "Paths" collection exists in our ArangoDB instance. The collection is built by the "Paths" service.

### Telemetry
First, we need to get telemetry up and running.

### Data Collection
Now, we aggregate the data and correlate fields from MDT and OpenBMP to the network.

### Bandwidth Calculation and Upsert
Finally, calculate the bandwidth. We then upsert that bandwidth into the Paths collection for the specific path. 
