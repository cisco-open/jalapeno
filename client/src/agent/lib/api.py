"""API wrapper for agent interaction with Voltron API."""
from ipaddress import ip_network
import requests

class API():

    voltron_ip = None
    voltron_port = None

    def __init__(self, voltron_ip, voltron_port):
        self.voltron_ip = voltron_ip
        self.voltron_port = voltron_port

    def query(self, src_ip, dest_ip, parameters):
        parameter_map = {
            'latency': self.latency_query
        }
        for parameter in parameters:
            if parameter not in parameter_map.keys():
                raise ValueError(
                    'No {0} parameter support'.format(parameter)
                )
            #TODO Reconcile the query labels either in client or service.
            return parameter_map[parameter](src_ip, dest_ip)

    def latency_query(self, src_ip, dest_ip):
        mutilate_destination = ip_network(
            '{ip}/{prefixlen}'.format(ip=dest_ip, prefixlen=24),
            False
        )
        dest_ip = '{ip}_{prefixlen}'.format(
            ip=str(mutilate_destination.network_address),
            prefixlen=mutilate_destination.prefixlen
        )
        params = {
            'router_src': src_ip,
            'prefix_dst': dest_ip,
            'weight_attribute': 'Latency',
            'default_weight': 60
        }
        return self.__send_query(params)

    def __send_query(self, params):
        api_base = self.__construct_api_base()
        return requests.get(
            api_base,
            params=params
        ).json()

    def __construct_api_base(self):
        return 'http://{0}:{1}/api/v1'.format(
            self.voltron_ip,
            self.voltron_port
        )
