"""Representation of a Router Interface"""
class RouterInterface(object):

    def __init__(self, node_id, interface, current_utilization):
        self.node_id = node_id
        self.interface = interface
        self.current_utilization = current_utilization

