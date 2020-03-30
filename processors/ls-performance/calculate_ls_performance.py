import time
import ls_topology_generator, upsert_ls_performance as db_upserter
from configs import influxconfig, arangoconfig
from telemetry_values import telemetry_value_mapper
from util import connections
from pyArango.connection import *

def main():
    influx_connection, arango_connection = connections.InfluxConn(), connections.ArangoConn()
    influx_client = influx_connection.connect_influx(influxconfig.host, influxconfig.port, influxconfig.user, influxconfig.password, influxconfig.dbname)
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)

    while(True):
        ls_topology = ls_topology_generator.generate_ls_topology(arango_client)
        for link_index in range(len(ls_topology)):
            current_ls_link = ls_topology[link_index]
            ls_topology_key = current_ls_link['key']
            router_ip = current_ls_link['RouterID']
            router_interface_ip = current_ls_link['InterfaceIP']
            router_hostname = ls_topology_generator.get_node_hostname(arango_client, router_ip)
            interface_name = collect_interface_name(influx_client, router_hostname, router_interface_ip)
            print("\nCalculating performance metrics for Link-State link out of %s(%s) through %s(%s)" % (router_ip, router_hostname,
                                                                                                          router_interface_ip, interface_name))
            calculated_performance_metrics = {}
            for telemetry_value, performance_metric in telemetry_value_mapper.items(): # the extended base-path and the value it represents
                current_performance_metric_dataset = collect_performance_dataset(influx_client, router_hostname, interface_name, telemetry_value)
                current_performance_metric_value = calculate_performance_metric_value(current_performance_metric_dataset)
                #print("Calculated %s: %s" % (performance_metric, current_performance_metric_value))
                calculated_performance_metrics[performance_metric] = current_performance_metric_value
            current_port_speed_dataset = collect_port_speed_dataset(influx_client, router_hostname, interface_name)
            current_port_speed_value = calculate_port_speed_value(current_port_speed_dataset)
            percent_util_inbound = calculated_performance_metrics["in-octets"]/float(((current_port_speed_value * 1000)/8))
            percent_util_outbound = calculated_performance_metrics["out-octets"]/float(((current_port_speed_value * 1000)/8))
            calculated_performance_metrics["speed"] = current_port_speed_value
            calculated_performance_metrics["percent-util-inbound"] = percent_util_inbound
            calculated_performance_metrics["percent-util-outbound"] = percent_util_outbound
            upsert_ls_performance(arango_client, ls_topology_key, calculated_performance_metrics)
            print("============================================================")
        time.sleep(30)

def collect_performance_dataset(influx_client, source, interface_name, telemetry_value):
    performance_metric_query = """SELECT non_negative_derivative(last(\"""" + telemetry_value + """\"), 5s)
    FROM \"openconfig-interfaces:interfaces/interface\"
    WHERE (\"source\" = '""" + source + """' AND \"name\" = '""" + interface_name + """')
    AND time >= now() - 5m GROUP BY time(200ms) fill(null);"""
    performance_metric_dataset = influx_client.query(performance_metric_query)
    return performance_metric_dataset

def calculate_performance_metric_value(performance_metric_dataset):
    rolling_avg = list(performance_metric_dataset.get_points())
    try:
        current_performance_metric_value = rolling_avg[-1]['non_negative_derivative']
    except IndexError:
        # no metrics found
        return 0
    return int(round(current_performance_metric_value))

def collect_interface_name(influx_client, source, interface_ip):
    map_query = """SELECT last(\"ip_information/ip_address\") AS \"interface_ip\", \"interface_name\"
    FROM \"Cisco-IOS-XR-pfi-im-cmd-oper:interfaces/interface-xr/interface\"
    WHERE (\"source\" = '""" + source + """') GROUP BY \"interface_name\";"""
    map = influx_client.query(map_query)
    map_list = list(map.get_points())
    interface_name = ""
    for index in range(len(map_list)):
        if((map_list[index]["interface_ip"] == interface_ip)):
            interface_name = map_list[index]["interface_name"]
    return interface_name

def collect_port_speed_dataset(influx_client, source, interface_name):
    port_speed_query = """SELECT last(\"speed\") AS \"speed\"
    FROM \"Cisco-IOS-XR-pfi-im-cmd-oper:interfaces/interface-xr/interface\"
    WHERE (\"source\" = '""" + source + """' AND \"interface_name\" = '""" + interface_name + """')
    AND time >= now() - 5m;"""
    port_speed_dataset = influx_client.query(port_speed_query)
    return port_speed_dataset

def calculate_port_speed_value(port_speed_dataset):
    speed_datapoints = list(port_speed_dataset.get_points())
    try:
        port_speed = speed_datapoints[-1]['speed']
    except IndexError:
        # no metrics found
        port_speed = 1
    return port_speed

def upsert_ls_performance(arango_client, ls_topology_key, performance_metrics):
    in_unicast_pkts, out_unicast_pkts = performance_metrics["in-unicast-pkts"], performance_metrics["out-unicast-pkts"]
    in_multicast_pkts, out_multicast_pkts = performance_metrics["in-multicast-pkts"], performance_metrics["out-multicast-pkts"]
    in_broadcast_pkts, out_broadcast_pkts = performance_metrics["in-broadcast-pkts"], performance_metrics["out-broadcast-pkts"]
    in_discards, out_discards = performance_metrics["in-discards"], performance_metrics["out-discards"]
    in_errors, out_errors = performance_metrics["in-errors"], performance_metrics["out-errors"]
    in_octets, out_octets = performance_metrics["in-octets"], performance_metrics["out-octets"]
    percent_util_inbound, percent_util_outbound = performance_metrics["percent-util-inbound"], performance_metrics["percent-util-outbound"]
    speed = performance_metrics["speed"]
    db_upserter.update_ls_performance(arango_client, ls_topology_key, in_unicast_pkts, out_unicast_pkts,
                                      in_multicast_pkts, out_multicast_pkts, in_broadcast_pkts, out_broadcast_pkts,
                                      in_discards, out_discards, in_errors, out_errors, in_octets, out_octets, speed,
                                      percent_util_inbound, percent_util_outbound)

if __name__ == '__main__':
    main()
