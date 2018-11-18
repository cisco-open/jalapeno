#! /usr/bin/env python
"""Database connection library.
Currently includes connection capabilities for ArangoDB and InfluxDB.
"""

from pyArango.connection import Connection
from influxdb import InfluxDBClient

class ArangoConn():
    """Connection class for ArangoDB."""
    def connect_arango(self, url, db_name, username, password):
        arango_connection = Connection(arangoURL=url, username=username, password=password)
        db_connection = arango_connection[db_name]
	return db_connection

    def aql_query(collection, aql, rawResults=True):
	query_result = collection.AQLQuery(aql, rawResults = rawResults)
	return query_result

class InfluxConn():
    """Connection class for InfluxDB."""
    def connect_influx(self, host, port, user, password, dbname):
        influx_client = InfluxDBClient(host, port, user, password, dbname)
        return influx_client

# other database connection capabilities can be added here
