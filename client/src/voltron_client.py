"""Voltron client"""
#!/usr/bin/env python
import os
import logging
import argparse
from lib import Agent

def main():
    setup_logging()
    args = setup_args()
    arango_api_gw = args.arango_api_gw.split(':')
    logging.info('Setting up Voltron Agent!')
    voltron_arango_agent = Agent(arango_api_gw[0], arango_api_gw[1])
    logging.info('Optimizing flow!')
    voltron_arango_agent.optimize_flow(
        args.src_ip,
        args.src_transport_ip,
        args.src_gw_ip,
        args.dest_ip,
        args.services
    )

def setup_args():
    parser = argparse.ArgumentParser(
        description="Voltron Demo Client Agent"
    )
    parser.add_argument('--context',
        nargs='?',
        help='host | router',
        default='host'
    )
    parser.add_argument('src_ip',
        help='source ip'
    )
    parser.add_argument('src_transport_ip',
        help='transport ip'
    )
    parser.add_argument('src_gw_ip',
        help='upstream gateway router ip'
    )
    parser.add_argument('dest_ip',
        help='destination ip'
    )

    parser.add_argument('arango_api_gw',
        help='host:port'
    )
    parser.add_argument('services',
        nargs='+',
        help='services of interest',
    )
    return parser.parse_args()


def setup_logging():
    logging.getLogger().setLevel(logging.INFO)
    logging.getLogger("requests").setLevel(logging.WARNING)

if __name__ == '__main__':
    main()
    exit(0)
