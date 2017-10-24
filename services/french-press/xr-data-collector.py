#!/usr/bin/env python
import logging
import multiprocessing
from kafka import KafkaConsumer
import json
import requests
import ConfigParser
import argparse

class Collector(multiprocessing.Process):
    daemon = True
    def run(self):
        parser = argparse.ArgumentParser()
        parser.add_argument('--kafkaEndpoint', type=str, help='The Kafka endpoint')
        parser.add_argument('--arangodbEndpoint', type=str, help='The ArangoDB endpoint')
        args = parser.parse_args()
        consumer = KafkaConsumer(group_id="voltronLatency", bootstrap_servers=args.kafkaEndpoint, auto_offset_reset='latest')
        consumer.subscribe(["voltron.Latency"])
        fd = open("router-prefixes.txt", "r")
        prefixes = fd.readlines()
        fd.close()
        while True:
            try:
                for message in consumer:
                    latency_data = json.loads(message.value)
                    from_ip = latency_data["from_ip"]
                    to_ip = latency_data["to_ip"]
                    for line in prefixes:
                        if to_ip in line:
                            to_ip = line.split(":")[1]  # Convert host IP address to the prefix
                            to_ip = to_ip.strip()
                    latency = latency_data["latency"]
                    res = requests.get('http://'+args.arangodbEndpoint+'/_db/voltron/queries/latency/'+from_ip+'/'+to_ip+'/'+latency)
                    if res.status_code==200:
                        print 'Inserted latency '+latency+' from '+from_ip+' to '+to_ip
                    else:
                        print 'Error inserting latency '+latency+' from '+from_ip+' to '+to_ip
            except Exception as e:
                print("Unable to parse latency data: {0}".format(e))
                continue

if __name__ == "__main__":
    logging.basicConfig(
        format='%(asctime)s.%(msecs)s:%(name)s:%(thread)d:%(levelname)s:%(process)d:%(message)s',
        level=logging.INFO
        )
    logging.getLogger("urllib3").setLevel(logging.WARNING)
    Collector().start()
