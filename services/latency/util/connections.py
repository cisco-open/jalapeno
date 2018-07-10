#! /usr/bin/env python
"""Database connection library.
Currently includes connection capabilities for ArangoDB.
"""

from pyArango.connection import Connection

class ArangoConn():
    """Connection class for ArangoDB."""
    def connect_arango(self, url, db_name, username, password):
        arango_connection = Connection(arangoURL=url, username=username, password=password)
        db_connection = arango_connection[db_name]
	return db_connection

    def aql_query(collection, aql, rawResults=True):
	query_result = collection.AQLQuery(aql, rawResults = rawResults)
	return query_result

# other database connection capabilities can be added here
