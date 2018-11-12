import time
import internal_links_generator, upsert_internal_links_performance as db_upserter
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
        internal_links = internal_links_generator.generate_internal_links(arango_client)
        for link in internal_links:
            router_ip = link['RouterIP']
            router_interface_ip = link['InterfaceIP']
            routerIP_interfaceIP = router_ip + "_" + router_interface_ip
            if telemetry_interface_mapper.get(routerIP_interfaceIP) != None:
                telemetry_interface_info = telemetry_interface_mapper[routerIP_interfaceIP]
                telemetry_producer, producer_interface = telemetry_interface_info[0], telemetry_interface_info[1] 	# i.e. r0.622 and Gig0/0/0/1
                print("\nCalculating performance metrics for internal-link out of %s(%s) through %s(%s)" % (telemetry_producer, router_ip, 
                                                                                                      producer_interface, router_interface_ip))
                calculated_performance_metrics = {}
                for telemetry_value, performance_metric in telemetry_value_mapper.iteritems(): # the extended base-path and the value it represents
                    current_performance_metric_dataset = collect_performance_dataset(influx_client, telemetry_producer, producer_interface, telemetry_value)
                    current_performance_metric_value = calculate_performance_metric_value(current_performance_metric_dataset)
                    print("Calculated %s: %s" % (performance_metric, current_performance_metric_value))
                    calculated_performance_metrics[performance_metric] = current_performance_metric_value
                upsert_internal_link_performance(router_ip, router_interface_ip, calculated_performance_metrics)
            else:
                print("\n%s (interface %s) is not configured to send telemetry" % (router_ip, router_interface_ip))
	time.sleep(30)

def create_collection(arango_client):
    """Create new collection in ArangoDB. If the collection exists, connect to that collection."""
    collection_name = queryconfig.collection  # the collection name is set in queryconfig
    print("Creating " + collection_name + " collection in Arango")
    try:
        collection = arango_client.createCollection(name=collection_name)
    except CreationError:
        print(collection_name + " collection already exists!")

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
   
def upsert_internal_link_performance(internal_router_ip, internal_router_interface, performance_metrics):
    in_unicast_pkts, out_unicast_pkts = performance_metrics["in-unicast-pkts"], performance_metrics["out-unicast-pkts"]
    in_multicast_pkts, out_multicast_pkts = performance_metrics["in-multicast-pkts"], performance_metrics["out-multicast-pkts"]
    in_broadcast_pkts, out_broadcast_pkts = performance_metrics["in-broadcast-pkts"], performance_metrics["out-broadcast-pkts"]
    in_discards, out_discards = performance_metrics["in-discards"], performance_metrics["out-discards"]
    in_errors, out_errors = performance_metrics["in-errors"], performance_metrics["out-errors"]
    internal_link_performance_key = internal_router_ip + "_" + internal_router_interface    
    db_upserter.upsert_internal_link_performance(internal_link_performance_key, internal_router_ip, internal_router_interface, in_unicast_pkts, out_unicast_pkts,
                                               in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts, out_broadcast_pkts, in_discards, out_discards, 
                                               in_errors, out_errors)
   
if __name__ == '__main__':
    main()
