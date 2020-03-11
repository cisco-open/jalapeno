# InternalLinks Performance Service

InternalLinks Performance Calculation Process 
(done for each "link"(interface) in the InternalRoutersInterface collection)
- collect telemetry data (specifically from OpenConfig data models)
- aggregate necessary fields from telemetry
- execute performance calculations on a rolling average (for example, out-unicast-pkts) 
- upsert the performance metrics in a InternalLink_Performance document and up into the InternalLinks_Performance collection.

## Details
An "InternalRoutersInterface" collection exists in our ArangoDB instance. The collection is built by the "Topology" service.

### Telemetry
First, we need to get telemetry up and running.

### Data Collection
Now, we aggregate the data and correlate fields from MDT and OpenBMP to the network.

### Performance Calculation and Upsert
Finally, calculate the various performance metrics. We then upsert that bandwidth into the InternalLinks_Performance collection. 
