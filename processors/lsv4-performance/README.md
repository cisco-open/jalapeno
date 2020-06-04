#LSv4-Performance Processor

LSv4-Performance Calculation Process 
(done for each "link"(interface) in the LSv4_Topology collection)
- collect telemetry data (specifically from OpenConfig data models)
- aggregate necessary fields from telemetry
- execute performance calculations on a rolling average (for example, out-unicast-pkts) 
- upsert the performance metrics in an LSv4_Topology document and up into the LSv4_Topology collection.

## Details
An "LSv4_Topology" collection exists in our ArangoDB instance. The collection is built by the "LSv4-Processor".

### Telemetry
First, we need to get telemetry up and running.

### Data Collection
Now, we aggregate the data and correlate fields from MDT and OpenBMP to the network.

### Performance Calculation and Upsert
Finally, calculate the various performance metrics. We then upsert that bandwidth into the "LSv4_Topology" collection. 
