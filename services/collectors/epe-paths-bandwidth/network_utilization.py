import argparse, json, time
import interface_linkedge
from router import RouterInterface
from upsert_bandwidth import UpsertBandwidth
from util import connections
from configs import influxconfig
from interface_linkedge import interface_linkedge_mapper
from bandwidth_queries import link_utilization_query

def main():
    db = connections.InfluxConn()
    client = db.connect_influx(influxconfig.host, influxconfig.port, influxconfig.user, influxconfig.password, influxconfig.dbname)
    while(True):
        print("Querying data: " + link_utilization_query)
        utilization_datasets = client.query(link_utilization_query)
        link_util_scores = calculate_link_utilization(utilization_datasets)
        upsert_link_utilization(link_util_scores)
        time.sleep(15)

def calculate_link_utilization(utilization_datasets):
    links = []
    for dataset in utilization_datasets:
        rolling_avg = list(dataset.get_points())
        current_utilization = rolling_avg[-1]
        producer = [dataset_key[1]['Producer'] for dataset_key in dataset.keys()]
        interface = [dataset_key[1]['interface-name'] for dataset_key in dataset.keys()]
        current_router_interface = RouterInterface(producer, interface, current_utilization)
        links.append(current_router_interface)
    return links

def upsert_link_utilization(link_util_scores):
    for link in link_util_scores:
        print link.node_id, link.interface, link.current_utilization
        try:
            arango_key = interface_linkedge_mapper[str(str(link.node_id[0]) + "_" + str(link.interface[0]))]
	    print("Upserting bandwidth")
            bandwidth_obj = UpsertBandwidth()
            bandwidth_obj.send_bandwidth(arango_key, link.current_utilization["moving_average"])
	    print("Upserted bandwidth")
        except:
	    print("No corresponding link edge for current interface utilization")
	    pass

if __name__ == '__main__':
    main()
