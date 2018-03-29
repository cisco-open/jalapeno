#! /usr/bin/env python
"""This class has the core functions to calculate and upsert bandwidth."""

from scapy.all import *
from pyArango.connection import *
from configs import arangoconfig, queryconfig
import logging, decimal
from util import connections

class UpsertBandwidth(object):
    def send_bandwidth(self, key, bandwidth):
	"""Update bandwidth in Paths collection for the given document specified by key."""
        connection = connections.ArangoConn()
        database = connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
        collection = database['Paths']
        aql = """FOR p in Paths
            FILTER p._key == @key
            UPDATE p with { Bandwidth : @bandwidth } in Paths """
        bindVars = {'key': key,'bandwidth': bandwidth}
        database.AQLQuery(aql, rawResults=True, bindVars=bindVars)

    def calculate_bandwidth(self, key):
	"""Calculate the highest available bandwidth for a specific path."""
	# bandwidth = some calculation here
        self.send_bandwidth(key, bandwidth)
