# EPEPaths Bandwidth OpenConfig Service

EPEPaths Bandwidth OpenConfig Calculation Process 
(done for each path in the EPEPaths collection)
- collect telemetry data (specifically from OpenConfig data models)
- aggregate necessary fields from telemetry
- execute bandwidth calculation on a rolling average (specifically out-unicast-pkts) 
- upsert the bandwidth in a EPEPath_Bandwidth_OpenConfig document and up into the EPEPaths_Bandwidth_OpenConfig collection

## Details
A "EPEPaths" collection exists in our ArangoDB instance. The collection is built by the "EPEPaths" service.

### Telemetry
First, we need to get telemetry up and running.

### Data Collection
Now, we aggregate the data and correlate fields from MDT and OpenBMP to the network.

### Bandwidth Calculation and Upsert
Finally, calculate the bandwidth. We then upsert that bandwidth into the EPEPaths_Bandwidth_OpenConfig collection for the specific path. 
