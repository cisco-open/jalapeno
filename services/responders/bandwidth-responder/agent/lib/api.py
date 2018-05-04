"""API wrapper for agent interaction with Arango API."""
from ipaddress import ip_network
import requests

class API():

    voltron_ip = None
    voltron_port = None

    def __init__(self, voltron_ip, voltron_port):
        self.voltron_ip = voltron_ip
        self.voltron_port = voltron_port

    def query(self, src_ip, src_gw_ip, dest_ip, parameters):
        parameter_map = {
            'bandwidth': self.bandwidth_query
        }
        for parameter in parameters:
            if parameter not in parameter_map.keys():
                raise ValueError(
                    'No {0} parameter support'.format(parameter)
                )
            return parameter_map[parameter](src_ip, src_gw_ip, dest_ip)

    def bandwidth_query(self, src_ip, src_gw_ip, dest_ip):
        mutilate_destination = ip_network(
            '{ip}/{prefixlen}'.format(ip=dest_ip, prefixlen=24),
            False
        )
        dest_ip = '{ip}_{prefixlen}'.format(
            ip=str(mutilate_destination.network_address),
            prefixlen=mutilate_destination.prefixlen
        )
        params = {
            'router_src': src_gw_ip,
            'prefix_dst': dest_ip,
            'weight_attribute': 'bandwidth',
        }
        labels = self.__send_query(params)
        return labels

    def __send_query(self, params):
        arango_api_call = self.__construct_arango_api(params)
        return requests.get(arango_api_call).json()

    def __construct_arango_api(self, params):
        return 'http://{0}:{1}/_db/voltron/queries/{2}/{3}/{4}'.format(
            self.voltron_ip,
            self.voltron_port,
            params['router_src'],
            params['prefix_dst'],
            params['weight_attribute']
        )

