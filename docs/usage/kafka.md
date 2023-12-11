# Kafka

To access Kafka and validate Jalapeno topics and data:

1. Access container:

    ```bash
    kubectl exec -it <kafka pod name> /bin/bash -n jalapeno
    ```

2. Change directory to `bin` and unset `JMX_PORT`

    ```bash
    cd bin
    unset JMX_PORT
    ```
  
3. List Jalapeno topics:

    ```bash
    ./kafka-topics.sh --zookeeper zookeeper.jalapeno.svc:2181 --list  
    ```

4. Listen for a topic's messages:

    ```bash
    ./kafka-console-consumer.sh --bootstrap-server localhost:9092  --topic gobmp.parsed.ls_node
    ./kafka-console-consumer.sh --bootstrap-server localhost:9092  --topic gobmp.parsed.ls_link
    ./kafka-console-consumer.sh --bootstrap-server localhost:9092  --topic gobmp.parsed.l3vpn_v4
    ```

If the topic provides JSON output similar to the below sample, we know that GoBMP is successfully writing BMP messages to Kafka.

???+ note
    There may not be active messages in Kafka's buffer at any given time.

    BMP messages can be triggered by either clearing BGP link-state or shut/no-shut on the router's BMP server.

Sample output:

```json
{
  "action": "add",
  "router_hash": "9e1a9a3663f25a297ed16a834b473eb0",
  "domain_id": 0,
  "router_ip": "10.0.0.10",
  "peer_hash": "308bc76d9c523fce904af3300c97d77e",
  "peer_ip": "10.0.0.12",
  "peer_asn": 65000,
  "timestamp": "Nov  3 20:28:02.000896",
  "igp_router_id": "0000.0000.0003",
  "router_id": "10.0.0.3",
  "asn": 65000,
  "mt_id": [
    0,
    2
  ],
  "isis_area_id": "49.0901",
  "protocol": "IS-IS Level 2",
  "protocol_id": 2,
  "node_flags": 0,
  "name": "R03-LSR",
  "ls_sr_capabilities": {
    "sr_capability_flags": 128,
    "sr_capability_tlv": [
      {
        "range": 64000,
        "sid_tlv": {
          "sid": "AYag"
        }
      }
    ]
  },
  "sr_algorithm": [
    0,
    1
  ],
  "sr_local_block": {
    "flags": 0,
    "subranges": [
      {
        "range_size": 1000,
        "label": 15000
      }
    ]
  },
  "srv6_capabilities_tlv": {
    "flag": 0
  },
  "node_msd": [
    {
      "msd_type": 1,
      "msd_value": 10
    }
  ],
  "isprepolicy": false,
  "is_adj_rib_in": false
}
```
