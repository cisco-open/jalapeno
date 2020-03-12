import time
import internal_links_generator, upsert_internal_link_performance as db_upserter
from configs import influxconfig, arangoconfig, queryconfig
from telemetry_interfaces import telemetry_interface_mapper
from telemetry_values import telemetry_value_mapper
from util import connections
from pyArango.connection import *

def main():
    influx_connection, arango_connection = connections.InfluxConn(), connections.ArangoConn()
    influx_client = influx_connection.connect_influx(influxconfig.host, influxconfig.port, influxconfig.user, influxconfig.password, influxconfig.dbname)
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    create_collections(arango_client)

    while(True):
        internal_links = internal_links_generator.generate_internal_links(arango_client)
        #print(internal_links)
        for link in range(len(internal_links)):
            #print(internal_links[link])
            router_ip = internal_links[link]['RouterIP']
            router_interface_ip = internal_links[link]['InterfaceIP']
            #print(router_ip, router_interface_ip)
            if internal_links[link]["key"] not in telemetry_interface_mapper:
                print("\n%s (interface %s) is not currently configured for internal telemetry measurements or calculations" % (router_ip, router_interface_ip))
                continue
            routerIP_interfaceIP = router_ip + "_" + router_interface_ip
            telemetry_interface_info = telemetry_interface_mapper[routerIP_interfaceIP]
            #print(routerIP_interfaceIP, telemetry_interface_info)
            telemetry_producer, producer_interface = telemetry_interface_info[0], telemetry_interface_info[1] 	# i.e. r0.622 and Gig0/0/0/1
            print("\nCalculating performance metrics for internal-link out of %s(%s) through %s(%s)" % (telemetry_producer, router_ip, 
                                                                                                      producer_interface, router_interface_ip))
            calculated_performance_metrics = {}
            #print(telemetry_producer, producer_interface)
            for telemetry_value, performance_metric in telemetry_value_mapper.items(): # the extended base-path and the value it represents
                current_performance_metric_dataset = collect_performance_dataset(influx_client, telemetry_producer, producer_interface, telemetry_value)
                current_performance_metric_value = calculate_performance_metric_value(current_performance_metric_dataset)
                #print("Calculated %s: %s" % (performance_metric, current_performance_metric_value))
                calculated_performance_metrics[performance_metric] = current_performance_metric_value
            current_port_speed_dataset = collect_port_speed_dataset(influx_client, telemetry_producer, producer_interface)
            current_port_speed_value = calculate_port_speed_value(current_port_speed_dataset)
            percent_util_inbound = calculated_performance_metrics["in-octets"]/float(((current_port_speed_value * 1000)/8))
            percent_util_outbound = calculated_performance_metrics["out-octets"]/float(((current_port_speed_value * 1000)/8))
            calculated_performance_metrics["speed"] = current_port_speed_value
            calculated_performance_metrics["percent-util-inbound"] = percent_util_inbound
            calculated_performance_metrics["percent-util-outbound"] = percent_util_outbound
        
            # There are two places to upsert these metrics: the InternalRouterInterfaces collection, and the InternalLinkEdges collection.
            internal_router_interface_key = router_ip + "_" + router_interface_ip
            upsert_internal_link_performance(internal_router_interface_key, calculated_performance_metrics, "InternalRouterInterfaces")
            internal_link_edges = internal_links_generator.generate_internal_link_edges(arango_client, router_ip, router_interface_ip)

            for internal_link_edge_index in range(len(internal_link_edges)):
                internal_link_edge = internal_link_edges[internal_link_edge_index]
                #print(internal_link_edge)
                #print(internal_link_edge['dest_intf_ip'])
                #print(internal_link_edge['destination'])
                #print(router_ip)
                #print(router_interface_ip)
                internal_link_edge_key = router_ip + "_" + router_interface_ip + "_" + internal_link_edge['dest_intf_ip'] + "_" + internal_link_edge['destination'].replace("Routers/", "")
                upsert_internal_link_performance(internal_link_edge_key, calculated_performance_metrics, "InternalLinkEdges")
            print("============================================================")
        time.sleep(30)


def create_collections(arango_client):
    """Create new collection in ArangoDB. If the collection exists, connect to that collection."""
    collections = {queryconfig.collection, queryconfig.edge_collection}  # the collection name is set in queryconfig
    for collection in collections:
        print("Creating " + collection + " collection in Arango")
        try:
            created_collection = arango_client.createCollection(name=collection)
        except CreationError:
            print(collection + " collection already exists!")

def collect_performance_dataset(influx_client, telemetry_producer, interface_name, telemetry_value):
    performance_metric_query = """SELECT non_negative_derivative(last(\"""" + telemetry_value + """\"), 5s)
    FROM \"openconfig-interfaces:interfaces/interface\"
    WHERE (\"Producer\" = '""" + telemetry_producer + """' AND \"name\" = '""" + interface_name + """')
    AND time >= now() - 5m GROUP BY time(200ms) fill(null);"""
    #print(performance_metric_query)
    performance_metric_dataset = influx_client.query(performance_metric_query)
    return performance_metric_dataset

def collect_port_speed_dataset(influx_client, telemetry_producer, interface_name):
    port_speed_query = """SELECT \"speed\"
    FROM \"Cisco-IOS-XR-pfi-im-cmd-oper:interfaces/interface-xr/interface\"
    WHERE (\"Producer\" = '""" + telemetry_producer + """' AND \"interface-name\" = '""" + interface_name + """')
    AND time >= now() - 5m;"""
    #print(port_speed_query)
    port_speed_dataset = influx_client.query(port_speed_query)
    return port_speed_dataset

def calculate_performance_metric_value(performance_metric_dataset):
    rolling_avg = list(performance_metric_dataset.get_points())
    current_performance_metric_value = rolling_avg[-1]['non_negative_derivative']
    return int(round(current_performance_metric_value))

def calculate_port_speed_value(port_speed_dataset):
    speed_datapoints = list(port_speed_dataset.get_points())
    port_speed = speed_datapoints[-1]['speed']
    return port_speed
   
def upsert_internal_link_performance(internal_link_performance_key, performance_metrics, collection_name):
    in_unicast_pkts, out_unicast_pkts = performance_metrics["in-unicast-pkts"], performance_metrics["out-unicast-pkts"]
    in_multicast_pkts, out_multicast_pkts = performance_metrics["in-multicast-pkts"], performance_metrics["out-multicast-pkts"]
    in_broadcast_pkts, out_broadcast_pkts = performance_metrics["in-broadcast-pkts"], performance_metrics["out-broadcast-pkts"]
    in_discards, out_discards = performance_metrics["in-discards"], performance_metrics["out-discards"]
    in_errors, out_errors = performance_metrics["in-errors"], performance_metrics["out-errors"]
    in_octets, out_octets = performance_metrics["in-octets"], performance_metrics["out-octets"]
    percent_util_inbound, percent_util_outbound = performance_metrics["percent-util-inbound"], performance_metrics["percent-util-outbound"]
    speed = performance_metrics["speed"]
    db_upserter.upsert_internal_link_performance(collection_name, internal_link_performance_key, in_unicast_pkts, out_unicast_pkts,
                                               in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts, out_broadcast_pkts,
                                               in_discards, out_discards, in_errors, out_errors, in_octets, out_octets, speed,
                                               percent_util_inbound, percent_util_outbound)
if __name__ == '__main__':
    main()