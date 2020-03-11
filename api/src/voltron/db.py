"""Module for database wrappers."""
from os import environ
from arango import ArangoClient


class ArangoDBConnection(object):
    """Wrapper for ArangoDB interaction. Capabilities exposed as necessary.
    Raw usage of underlying library is available through client or db attributes.
    """

    def __init__(self, protocol=None, host=None, port=None, database=None, username=None, password=None):
        """Construct wrapper, optionally but recommended specifying database at start."""
        self.protocol = protocol or environ.get('ARANGODB_PROTOCOL')
        self.host = host or environ.get('ARANGODB_HOST')
        self.port = port or environ.get('ARANGODB_PORT')
        if not all({self.protocol, self.host, self.port}):
            raise ValueError('protocol, host, and port must be specified directly or via env variables!')
        self.client = ArangoClient(protocol=self.protocol, host=self.host, port=self.port)
        self.database = database or environ.get('ARANGODB_DB')
        username = username or environ.get('ARANGODB_USER')
        password = password or environ.get('ARANGODB_PASSWORD')
        if self.database:
            self.db = self.select_database(self.database, username, password)
    
    def select_database(self, database, username, password):
        """Return the underlying library database selection."""
        return self.client.db(database, username, password)
    
    def query_aql(self, aql, bind_vars=None):
        """Execute an AQL query and return the results."""
        return self.db.aql.execute(aql, bind_vars=bind_vars)

    def __repr__(self):
        return '%s|%s' % (str(self.host), str(self.database))
