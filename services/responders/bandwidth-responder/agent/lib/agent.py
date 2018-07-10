"""Bandwidth Responder agent accessing Arango API"""
import time
import requests
from . import log
from .flow import Flow
from .api import API
from .exceptions import FlowRouteException

class Agent():
    api = None
    flows = None

    def __init__(self, voltron_ip, voltron_port):
        self.api = API(voltron_ip, voltron_port)

    def get_bandwidth_labels(self, src_ip, src_transport_ip, src_gw_ip, dest_ip, parameters):
        new_flow = Flow(src_ip, src_transport_ip, src_gw_ip, dest_ip, parameters)
        label_stack = self.api.query(
            new_flow.src_ip,
            new_flow.src_gw_ip,
            new_flow.dest_ip,
            new_flow.parameters
        )
        return label_stack
