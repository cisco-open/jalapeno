# Jalapeno Infrastructure: Telemetry

## Deployment

To deploy Jalapeno's telemetry infrastructure, configure `hosts.json` to reflect which devices you would like telemetry data to be streamed from.

Once your hosts file is filled, run:
```
python deploy_telemetry.py
```

That deployment script configures telemetry according to device information and the templates in 'config_xr_template' and 'config_xe_template'

Note: `hosts.json.example` is included as an example of how `hosts.json` should look.

In the event you need to remove	Pipeline or reassess your telemetry configuration, run:
```
python remove_telemetry
```

## Helpful Hints

#### To confirm an active telemetry subscription: 
SSH onto your network device. Run:
```
show telemetry model-driven subscription
```

#### To confirm Pipeline is receiving data from your network devices:
Enter the Pipeline-Ingress pod (through CLI or the OpenShift UI). Run:
```
cat pipeline.log
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
