"""Utility functions for telemetry automation."""
import json
import logging

def load_dict_from_json(json_input, raw=False):
    json_dict = None
    if raw:
        json_dict = json.loads(json_input)
    else:
        with open(json_input, 'r') as json_fd:
            json_dict = json.load(json_fd)
    return json_dict

def get_hosts(hosts_file='hosts.json'):
    """Loads hosts from file."""
    logging.debug('Loading hosts from %s', hosts_file)
    return load_dict_from_json(hosts_file)
