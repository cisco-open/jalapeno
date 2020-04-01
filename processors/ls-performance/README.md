#LS-Performance Processor

LS-Performance Calculation Process 
(done for each "link"(interface) in the LS_Topology collection)
- collect telemetry data (specifically from OpenConfig data models)
- aggregate necessary fields from telemetry
- execute performance calculations on a rolling average (for example, out-unicast-pkts) 
- upsert the performance metrics in an LS_Topology document and up into the LS_Topology collection.

## Details
An "LS_Topology" collection exists in our ArangoDB instance. The collection is built by the "LS-Processor".

### Telemetry
First, we need to get telemetry up and running.

### Data Collection
Now, we aggregate the data and correlate fields from MDT and OpenBMP to the network.

### Performance Calculation and Upsert
Finally, calculate the various performance metrics. We then upsert that bandwidth into the "LS_Topology" collection. 
