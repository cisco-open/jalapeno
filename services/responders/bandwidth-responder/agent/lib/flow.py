"""Representation of a network flow for a route purpsoe.
Does not support complex tuple based routing, only simple source and dest IP.
"""

class Flow():

    src_ip = None
    src_gw_ip = None
    src_transport_ip = None
    dest_ip = None
    parameters = None
    label_stack = None

    def __init__(self, src_ip, src_transport_ip, src_gw_ip, dest_ip, parameters, label_stack=[]):
        self.src_ip = src_ip
        self.src_transport_ip = src_transport_ip
        self.src_gw_ip = src_gw_ip
        self.dest_ip = dest_ip
        self.parameters = parameters
        self.label_stack = label_stack

    def __eq__(self, other):
        return self.src_ip == other.src_ip and self.dest_ip == other.dest_ip
