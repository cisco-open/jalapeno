#! /usr/bin/env python
"""This class upserts bandwidth into ArangoDB"""
from scapy.all import *
from pyArango.connection import *
from configs import arangoconfig, queryconfig
import logging, decimal
from util import connections

class UpsertBandwidth(object):
    def send_bandwidth(self, key, bandwidth):
	"""Update bandwidth in LinkEdgesV4 collection for the given document specified by key."""
        connection = connections.ArangoConn()
        database = connection.connect_arango(arangoconfig.url, arangoconfig.database, arangoconfig.username, arangoconfig.password)
        collection = database['LinkEdgesV4']
        aql = """FOR e in LinkEdgesV4
            FILTER e._key == @key
            UPDATE e with { Bandwidth : @bandwidth } in LinkEdgesV4 """
        bindVars = {'key': key,'bandwidth': bandwidth}
        database.AQLQuery(aql, rawResults=True, bindVars=bindVars)

