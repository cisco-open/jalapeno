#!/usr/bin/env python
import logging
import multiprocessing
import ast
import json
import requests
import ConfigParser
import argparse
import sys
import os
import client
import datetime
from pprint import pprint
from kafka import KafkaConsumer


class Framework(object):
    """ Example class that calls the framework to update&
        query edges/register/heartbeat."""
    field = "latency"
    collector = {"name": field, "edgeType": "PrefixEdges", "fieldName": field}

    def __init__(self, framework_ep):
        self.last_heartbeat = None
        if "http" not in framework_ep:
            framework_ep = "http://" + framework_ep
        if "v1" not in framework_ep:
            framework_ep += "/v1"
        api_client = client.ApiClient(host=framework_ep)
        self.client = client.DefaultApi(api_client=api_client)

    def register(self):
        """Collectors MUST attempt to register. The body object must contain
           name, edgeType and fieldName
        """
        try:
            self.client.add_collector(body=self.collector)
        except client.rest.ApiException as e:
            if e.status > 208:
                print "Could not register with framework. Exiting."
                return False
        return True

    def heartbeat(self):
        now = datetime.datetime.now()
        if self.last_heartbeat is None or ((now - self.last_heartbeat).seconds > 60):
            # Provided a callback makes this call async.
            # the lambda below is a no-op.
            self.client.heartbeat_collector(self.field, callback=lambda arg: True)
            self.last_heartbeat = now

    def update(self, _from, to, latency):
        """ We're given an internal router's interface ip.
            Get the edge from that ip. The "to" is the edge router.
            Use this to get/update the prefixEdge.
        """
        # Get LinkEdge where FromIP=_from
        edge = self.client.get_edge("LinkEdgesV4", "FromIP", _from)
        if "Prefixes" not in to:
            to = "Prefixes_" + to
        # Update PrefixEdges.Latency{with from/to} with latency value.
        es = client.EdgeScore(_from=edge["_to"], to=to, value=float(latency))
        return self.client.upsert_field("PrefixEdges", self.field, es)


class Collector(multiprocessing.Process):
    daemon = True

    def __init__(self, framework_client, kafka_ep,
                 prefix_file="router-prefixes.txt", group_id="voltronLatency",
                 topic="voltron.Latency"):
        self.kafka_endpoint = kafka_ep
        self.topic = topic
        self.group_id = group_id
        with open(prefix_file, 'r') as prefix_fd:
            prefixes = prefix_fd.readlines()[1:]

        self.prefix_map = {k: v.strip() for (k, v) in [p.split(":") for p in prefixes]}
        self.framework = framework_client

    def run(self):
        consumer = KafkaConsumer(group_id=self.group_id,
                                 bootstrap_servers=self.kafka_endpoint,
                                 auto_offset_reset='latest')
        consumer.subscribe([self.topic])
        while True:
            try:
                for message in consumer:
                    latency_data = json.loads(message.value)
                    from_ip = latency_data["from_ip"]
                    to_ip = latency_data["to_ip"]
                    if to_ip in self.prefix_map:
                        to_ip = self.prefix_map[to_ip]
                    latency = latency_data["latency"]
                    self.framework.update(from_ip, to_ip, latency)
                    print 'Inserted latency '+latency+' from '+from_ip+' to '+to_ip
                    self.framework.heartbeat()
            except Exception as e:
                print("Unable to parse latency data: {0}".format(e))
                continue

if __name__ == "__main__":
    logging.basicConfig(
        format='%(asctime)s.%(msecs)s:%(name)s:%(thread)d:%(levelname)s:%(process)d:%(message)s',
        level=logging.INFO
        )
    logging.getLogger("urllib3").setLevel(logging.WARNING)
    parser = argparse.ArgumentParser()
    parser.add_argument('--kafkaEndpoint', required=True, type=str, help='The Kafka endpoint')
    parser.add_argument('--frameworkEndpoint', required=True, type=str, help='The voltron framework endpoint')
    args = parser.parse_args()
    framework = Framework(args.frameworkEndpoint)
    if not framework.register():
        sys.exit(1)
    collector = Collector(framework, args.kafkaEndpoint).run()
