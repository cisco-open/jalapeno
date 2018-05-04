"""Voltron client."""
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

    # temporarily removing API gateway and using Arango API gateway until decision on which to use
    '''api_gw = args.api_gw.split(':')
    voltron_agent = Agent(api_gw[0], api_gw[1])
    voltron_agent.optimize_flow(
        args.src_ip,
        args.src_transport_ip,
        args.src_gw_ip,
        args.dest_ip,
        args.services
    )'''

    #logging.info('Running ping!')
    #os.system(
    #    'ping -I {src_ip} {dest_ip}'.format(
    #        src_ip=args.src_ip, dest_ip=args.dest_ip
    #    )
    #)

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

    ### temporarily removing Voltron API gateway until we decide whether or not we can use Arango API
    #parser.add_argument('api_gw',
    #    help='host:port'
    #)

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
