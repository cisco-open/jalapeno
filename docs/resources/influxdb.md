# Sample InfluxDB Queries

This section will cover some example queries that can be run against the InfluxDB instance.

## Get Interface Names & IPs

Provide all of Router 16's interface names and IPv4 addresses:

```text
SELECT last("ip_information/ip_address") FROM "Cisco-IOS-XR-pfi-im-cmd-oper:interfaces/interface-xr/interface" WHERE ("source" = 'R16-LSR') GROUP BY "interface_name"
```

## Get Interface IDs

Provide Router 16's interface IDs or indexes:

```text
SELECT last("if_index") FROM "Cisco-IOS-XR-pfi-im-cmd-oper:interfaces/interface-xr/interface" WHERE ("source" = 'R16-LSR') GROUP BY "interface_name"
```

## Get TX/RX Bytes

Provide transmit and receive bytes collected for a given router interface over the last hour (30 second collection interval)

```text
SELECT last("state/counters/out_octets"), last("state/counters/in_octets") FROM "openconfig-interfaces:interfaces/interface" WHERE ("name" = 'GigabitEthernet0/0/0/0' AND "source" = 'R12-LSR') AND time >= now() - 30m  GROUP BY time(30s) fill(null)
```

## Get MPLS Bytes

Provide total MPLS label switched bytes for a given interface or label value

```text
SELECT last("label_information/tx_bytes") FROM "Cisco-IOS-XR-fib-common-oper:mpls-forwarding/nodes/node/label-fib/forwarding-details/forwarding-detail" WHERE ("source" = 'R12-LSR' AND "label_information/outgoing_interface" = 'Gi0/0/0/4')
```

```text
SELECT last("label_information/label_information_detail/transmit_number_of_bytes_switched") FROM "Cisco-IOS-XR-fib-common-oper:mpls-forwarding/nodes/node/label-fib/forwarding-details/forwarding-detail" WHERE ("source" = 'R12-LSR' AND "label_value" = '100014')

```

## Get SR Traffic

Segment Routing Traffic Matrix collection

```text
SELECT last("base_counter_statistics/count_history/transmit_number_of_bytes_switched") FROM "Cisco-IOS-XR-infra-tc-oper:traffic-collector/vrf-table/default-vrf/afs/af/counters/prefixes/prefix" WHERE ("source" = 'R08-ABR' AND "label" = '100014') 
```
