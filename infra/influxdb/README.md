# InfluxDB

InfluxDB is Jalapeno's time-series database.

Network devices are configured to send performance data to Kafka. There, the data is parsed using Telegraf (a telemetry consumer) and stored in InfluxDB. The data stored in InfluxDB can be queried by Jalapeno Processors.

These queries construct and derive relevant metrics that inform Jalapeno's API Gateway responses to client requests. For example, the [Demo LS-Performance Processor](https://github.com/jalapeno/demo-processors/tree/main/lsv4-perf) generates rolling-averages of bytes sent out of a router's interface -- thus simulating link utilization. Those calculated insights are then inserted into ArangoDB and associated with their corresponding edges, providing a holistic view of the current state of the network.

Using InfluxDB as a historical data-store, trends can also be inferred based on historical analysis. Processors and applications can determine whether instantaneous measurements are extreme anomalies, and can enable requests for any number of threshold-based reactions. 

## Deploying InfluxDB
InfluxDB is deployed using `kubectl`, as seen in the [deploy_infrastructure script](../deploy_infrastructure.sh). The configurations for InfluxDB's deployment are in the various YAML files in the [jalapeno/infra/influxdb/](.) directory.

## Accessing InfluxDB
To access InfluxDB via Kubernetes, enter the pod's terminal and run:
```
influx
auth root jalapeno
use mdt_db
show series  // provides a list of all time-series in the mdt_db
```
To start working with Grafana dashboards see:
https://github.com/jalapeno/jalapeno/blob/master/infra/grafana/README.md

#### Sample InfluxDB Queries:
Provide all of Router 16's interface names and IPv4 addresses:
```
SELECT last("ip_information/ip_address") FROM "Cisco-IOS-XR-pfi-im-cmd-oper:interfaces/interface-xr/interface" WHERE ("source" = 'R16-LSR') GROUP BY "interface_name"
```

Provide Router 16's interface IDs or indexes:
```
SELECT last("if_index") FROM "Cisco-IOS-XR-pfi-im-cmd-oper:interfaces/interface-xr/interface" WHERE ("source" = 'R16-LSR') GROUP BY "interface_name"
```

Provide transmit and receive bytes collected for a given router interface over the last hour (30 second collection interval):
```
SELECT last("state/counters/out_octets"), last("state/counters/in_octets") FROM "openconfig-interfaces:interfaces/interface" WHERE ("name" = 'GigabitEthernet0/0/0/0' AND "source" = 'R12-LSR') AND time >= now() - 30m  GROUP BY time(30s) fill(null)
```

Provide total MPLS label switched bytes for a given interface or label value
```
SELECT last("label_information/tx_bytes") FROM "Cisco-IOS-XR-fib-common-oper:mpls-forwarding/nodes/node/label-fib/forwarding-details/forwarding-detail" WHERE ("source" = 'R12-LSR' AND "label_information/outgoing_interface" = 'Gi0/0/0/4')

SELECT last("label_information/label_information_detail/transmit_number_of_bytes_switched") FROM "Cisco-IOS-XR-fib-common-oper:mpls-forwarding/nodes/node/label-fib/forwarding-details/forwarding-detail" WHERE ("source" = 'R12-LSR' AND "label_value" = '100014') 
```

