import time
import epe_paths_generator, upsert_epe_paths_bandwidth_openconfig as db_upserter
from configs import influxconfig, arangoconfig, queryconfig
from telemetry_interfaces import telemetry_interface_mapper
from util import connections
from pyArango.connection import *

def main():
    influx_connection, arango_connection = connections.InfluxConn(), connections.ArangoConn()
    influx_client = influx_connection.connect_influx(influxconfig.host, influxconfig.port, influxconfig.user, influxconfig.password, influxconfig.dbname)
    arango_client = arango_connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
    create_collection(arango_client)

    while(True):
        epe_paths = epe_paths_generator.generate_epe_paths(arango_client)
        for telemetry_interface, router_info in telemetry_interface_mapper.iteritems():
            print("\nCalculating bandwidth usage for all destinations out of %s (router %s)" % (telemetry_interface, ' through interface '.join(router_info)))
            telemetry_interface_info = telemetry_interface.split('_')
            telemetry_producer, producer_interface = telemetry_interface_info[0], telemetry_interface_info[1]
            egress_router, egress_router_interface = router_info[0], router_info[1]
            current_link_utilization_dataset = collect_link_utilization_dataset(influx_client, telemetry_producer, producer_interface)
            link_util_score = calculate_link_utilization(current_link_utilization_dataset)
            matching_epe_paths = [path for path in epe_paths if path["Egress_Peer"] == egress_router and path["Egress_Interface"] == egress_router_interface]
            for epe_path in matching_epe_paths: 
                upsert_link_utilization(egress_router, egress_router_interface, link_util_score, epe_path)
        time.sleep(30)

def create_collection(arango_client):
    """Create new collection in ArangoDB. If the collection exists, connect to that collection."""
    collection_name = queryconfig.collection  # the collection name is set in queryconfig
    print("Creating " + collection_name + " collection in Arango")
    try:
        collection = arango_client.createCollection(name=collection_name)
    except CreationError:
        print(collection_name + " collection already exists!")

def collect_link_utilization_dataset(influx_client, telemetry_producer, interface_name):
    link_utilization_query = """SELECT moving_average(last(\"subinterfaces__subinterface__state__counters__out-unicast-pkts\"), 5)
    FROM \"openconfig-interfaces:interfaces/interface\"
    WHERE (\"Producer\" = '""" + telemetry_producer + """' AND \"name\" = '""" + interface_name + """')
    AND time >= now() - 5m GROUP BY time(200ms) fill(null);"""
    link_utilization_dataset = influx_client.query(link_utilization_query)
    return link_utilization_dataset

def calculate_link_utilization(link_utilization_dataset):
    rolling_avg = list(link_utilization_dataset.get_points())
    current_utilization = rolling_avg[-1]['moving_average']
    return current_utilization
   
def upsert_link_utilization(egress_router, egress_router_interface, link_util_score, epe_path):
    destination, labels = epe_path["Destination"], epe_path["Label_Path"]
    epe_paths_bandwidth_openconfig_key = "EPEPath:" + egress_router + "_" + egress_router_interface + "_" + destination    
    db_upserter.upsert_epe_path_bandwidth_openconfig(epe_paths_bandwidth_openconfig_key, egress_router, egress_router_interface, labels, destination, link_util_score)
   
if __name__ == '__main__':
    main()
