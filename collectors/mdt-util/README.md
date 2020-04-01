# Model-Driven Telemetry Utilities

## Configuring Telemetry

To deploy Jalapeno's telemetry infrastructure, configure `hosts.json` to reflect which devices you would like telemetry data to be streamed from.
To configure telemetry across multiple devices at once:
* Fill out `hosts.json` to reflect which devices you would like telemetry data to be streamed from
* Fill out the `<server_ip>` field in the `mdt_xr_config` file to point to the Jalapeno Cluster
* Run `python configure_telemetry.py`

That script uses netmiko to authenticate against the credentials supplied in `hosts.json` and deploys the config in `mdt_xr_config`.

Note: `hosts.json.example` is included as an example of how `hosts.json` should look.

The deployed telemetry configuration can also be removed using:
```
python remove_telemetry
```

## Helpful Hints

#### To confirm an active telemetry subscription: 
SSH onto your network device. Run:
```
show telemetry model-driven subscription
```

#### To confirm Kafka is receiving telemetry data: 
Enter the Zookeeper pod (through CLI or the OpenShift UI). Run:
```
cd /opt/kafka/bin
./kafka-topics.sh --zookeeper zookeeper.jalapeno.svc:2181 --list (lists topics)
./kafka-console-consumer.sh --zookeeper zookeeper.jalapeno.svc:2181 --topic jalapeno.telemetry --from-beginning (consumes and displays data in topic specified)
```

#### To see if Jalapeno is receiving telemetry data from a specific device:
Enter the InfluxDB pod (through CLI or the OpenShift UI). Run:
```
influx
auth jalapeno jalapeno
use mdt_db
show series
```
