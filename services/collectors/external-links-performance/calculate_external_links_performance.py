import time
import external_links_generator, upsert_external_link_performance as db_upserter
from configs import influxconfig, arangoconfig, queryconfig
from telemetry_interfaces import telemetry_interface_mapper
from telemetry_values import telemetry_value_mapper
from util import connections
from pyArango.connection import *

def main():
    influx_connection, arango_connection = connections.InfluxConn(), connections.ArangoConn()
    influx_client = influx_connection.connect_influx(influxconfig.host, influxconfig.port, influxconfig.user, influxconfig.password, influxconfig.dbname)
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    create_collection(arango_client)

    while(True):
        external_links = external_links_generator.generate_external_links(arango_client)
        for link in external_links:
            if link["key"] not in telemetry_interface_mapper:
                print("\n%s is not currently configured for telemetry" % str(link["RouterIP"]))
                continue
            router_ip = link['RouterIP']
            router_interface_ip = link['InterfaceIP']
            routerIP_interfaceIP = router_ip + "_" + router_interface_ip
            telemetry_interface_info = telemetry_interface_mapper[routerIP_interfaceIP]

            telemetry_producer, producer_interface = telemetry_interface_info[0], telemetry_interface_info[1] 	# i.e. r0.622 and Gig0/0/0/1
            print("\nCalculating performance metrics for external-link out of %s(%s) through %s(%s)" % (telemetry_producer, router_ip, 
                                                                                                      producer_interface, router_interface_ip))
            calculated_performance_metrics = {}
            for telemetry_value, performance_metric in telemetry_value_mapper.iteritems(): # the extended base-path and the value it represents
                current_performance_metric_dataset = collect_performance_dataset(influx_client, telemetry_producer, producer_interface, telemetry_value)
                current_performance_metric_value = calculate_performance_metric_value(current_performance_metric_dataset)
                print("Calculated %s: %s" % (performance_metric, current_performance_metric_value))
                calculated_performance_metrics[performance_metric] = current_performance_metric_value

            # There are two places to upsert these metrics: the BorderRouterInterfaces collection, and the ExternalLinkEdges collection.
            border_router_interface_key = router_ip + "_" + router_interface_ip
            upsert_external_link_performance(border_router_interface_key, calculated_performance_metrics, "BorderRouterInterfaces")
            external_link_edges = external_links_generator.generate_external_link_edges(arango_client, router_ip, router_interface_ip)
            for external_link_edge in external_link_edges:
                external_link_edge_key = router_ip + "_" + router_interface_ip + "_" + external_link_edge['dest_intf_ip'] + "_" + external_link_edge['destination'].replace("Routers/", "")
                upsert_external_link_performance(external_link_edge_key, calculated_performance_metrics, "ExternalLinkEdges")
	time.sleep(30)

def create_collection(arango_client):
    """Create new collection in ArangoDB. If the collection exists, connect to that collection."""
    collections = {queryconfig.collection, queryconfig.edge_collection}  # the collection name is set in queryconfig
    for collection in collections:
        print("Creating " + collection + " collection in Arango")
        try:
            created_collection = arango_client.createCollection(name=collection)
        except CreationError:
            print(collection + " collection already exists!")

def collect_performance_dataset(influx_client, telemetry_producer, interface_name, telemetry_value):
    performance_metric_query = """SELECT moving_average(last(\"""" + telemetry_value + """\"), 5)
    FROM \"openconfig-interfaces:interfaces/interface\"
    WHERE (\"Producer\" = '""" + telemetry_producer + """' AND \"name\" = '""" + interface_name + """')
    AND time >= now() - 5m GROUP BY time(200ms) fill(null);"""
    #print(performance_metric_query)
    performance_metric_dataset = influx_client.query(performance_metric_query)
    return performance_metric_dataset

def calculate_performance_metric_value(performance_metric_dataset):
    rolling_avg = list(performance_metric_dataset.get_points())
    current_performance_metric_value = rolling_avg[-1]['moving_average']
    return current_performance_metric_value
   
def upsert_external_link_performance(external_link_performance_key, performance_metrics, collection_name):
    in_unicast_pkts, out_unicast_pkts = performance_metrics["in-unicast-pkts"], performance_metrics["out-unicast-pkts"]
    in_multicast_pkts, out_multicast_pkts = performance_metrics["in-multicast-pkts"], performance_metrics["out-multicast-pkts"]
    in_broadcast_pkts, out_broadcast_pkts = performance_metrics["in-broadcast-pkts"], performance_metrics["out-broadcast-pkts"]
    in_discards, out_discards = performance_metrics["in-discards"], performance_metrics["out-discards"]
    in_errors, out_errors = performance_metrics["in-errors"], performance_metrics["out-errors"]
    in_octets, out_octets = performance_metrics["in-octets"], performance_metrics["out-octets"]
    db_upserter.upsert_external_link_performance(collection_name, external_link_performance_key, in_unicast_pkts, out_unicast_pkts,
                                               in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts, out_broadcast_pkts, 
                                               in_discards, out_discards, in_errors, out_errors, in_octets, out_octets)
   
if __name__ == '__main__':
    main()
