"""Main Voltron interaction class.
Serves as the host agent implementation to control host flows.
Currently only supports Linux.
"""

import threading
import subprocess
import time
import atexit
from . import log
from .flow import Flow
from .api import API
from .exceptions import FlowRouteException

class Agent():

    api = None
    flows = None
    flow_lock = None
    flow_monitor_thread = None
    running = None

    def __init__(self, voltron_ip, voltron_port):
        self.api = API(voltron_ip, voltron_port)
        self.flows = []
        self.flow_lock = threading.Lock()
        self.running = threading.Event()
        self.flow_monitor_thread = threading.Thread(target=self.__flow_monitor)
        self.flow_monitor_thread.start()
        atexit.register(self.cleanup_host)
        #TODO Clealy exit thread on exception/shutdown.

    def cleanup_host(self):
        log.info('Stopping Voltron.')
        self.running.set()
        self.flow_monitor_thread.join()
        log.info('Removing Voltron routes.')
        for flow in self.flows:
            self.__rm_flow_route(flow)
        log.info('Voltron has shut down.')

    def optimize_flow(self, src_ip, src_transport_ip, src_gw_ip, dest_ip, parameters):
        new_flow = Flow(src_ip, src_transport_ip, src_gw_ip, dest_ip, parameters)
        with self.flow_lock:
            if new_flow in self.flows:
                log.error('Flow is already being optimized!')
                raise ValueError('Flow is already being optimized!')
            self.flows.append(new_flow)
            self.__check_flow(new_flow)

    def __get_flow_labels(self, flow):
        label_stack = self.api.query(
            flow.src_gw_ip,
            flow.dest_ip,
            flow.parameters
        )
        return label_stack

    def __set_flow_route(self, flow, label_stack):
        parsed_label_stack = '/'.join(map(str, label_stack))
        route_cmd = 'ip route add {dest_ip}/32 encap mpls {label_stack} via inet {src_transport_ip} src {src_ip}'.format(
            src_ip=flow.src_ip,
            dest_ip=flow.dest_ip,
            src_transport_ip=flow.src_transport_ip,
            label_stack=parsed_label_stack
        )
        log.debug(route_cmd)
        cmd_ret = self.__exec_system_command(route_cmd)
        if cmd_ret:
            log.debug(cmd_ret)
            raise FlowRouteException(
                route_cmd,
                'Error setting flow route!'
            )
        flow.label_stack = label_stack
        log.info('Added route to {dest_ip}/32.'.format(dest_ip=flow.dest_ip))

    def __rm_flow_route(self, flow):
        route_cmd = 'ip route del {dest_ip}/32'.format(
            dest_ip=flow.dest_ip
        )
        log.debug(route_cmd)
        cmd_ret = self.__exec_system_command(route_cmd)
        if cmd_ret:
            log.debug(cmd_ret)
            raise FlowRouteException(
                route_cmd,
                'Error removing flow route!'
            )
        flow.label_stack = []
        log.info('Deleted route to {dest_ip}/32.'.format(dest_ip=flow.dest_ip))

    def __check_flow(self, flow):
        label_stack = self.__get_flow_labels(flow)
        if label_stack == flow.label_stack:
            return
        log.info(
            '{src_ip} -> {dest_ip}: {label_stack}'.format(
                src_ip=flow.src_ip,
                dest_ip=flow.dest_ip,
                label_stack=label_stack
            )
        )
        if flow.label_stack:
            self.__rm_flow_route(flow)
        self.__set_flow_route(flow, label_stack)

    def __flow_monitor(self):
        log.info('Starting flow monitor.')
        while not self.running.wait(3):
            #TODO Better check() logic for recent additions
            with self.flow_lock:
                for flow in self.flows:
                    self.__check_flow(flow)

    def __exec_system_command(self, command):
        return subprocess.check_output(
            command, shell=True
        ).decode('utf-8').strip()
