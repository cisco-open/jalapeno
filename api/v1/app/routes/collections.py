from fastapi import APIRouter, HTTPException
from arango import ArangoClient
from ..config.settings import Settings
from typing import Optional, List

router = APIRouter()
settings = Settings()

KNOWN_COLLECTIONS = {
    'graphs': [
        'ipv4_graph',
        'ipv6_graph',
        'igpv4_graph',
        'igpv6_graph'
    ],
    'prefixes': [
        'ebgp_prefix_v4',
        'ebgp_prefix_v6'
    ],
    'peers': [
        'bgp_node',
        'igp_node'
    ]
}

def get_db():
    client = ArangoClient(hosts=settings.database_server)
    try:
        db = client.db(
            settings.database_name,
            username=settings.username,
            password=settings.password
        )
        return db
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Could not connect to database: {str(e)}"
        )
@router.get("/collections")
async def get_collections(filter_graphs: Optional[bool] = None):
    """
    Get a list of collections in the database
    Optional: filter_graphs parameter:
    - None (default): show all collections
    - True: show only graph collections
    - False: show only non-graph collections
    """
    try:
        db = get_db()
        # Get all collections
        collections = db.collections()
        
        # Filter out system collections (those starting with '_')
        # Then apply graph filter if specified
        user_collections = [
            {
                'name': c['name'],
                'type': c['type'],
                'status': c['status'],
                'count': db.collection(c['name']).count()
            }
            for c in collections
            if not c['name'].startswith('_') and 
               (filter_graphs is None or  # Show all if no filter
                (filter_graphs and c['name'].endswith('_graph')) or  # Only graphs
                (not filter_graphs and not c['name'].endswith('_graph')))  # Only non-graphs
        ]
        
        # Sort by name
        user_collections.sort(key=lambda x: x['name'])
        
        return {
            'collections': user_collections,
            'total_count': len(user_collections),
            'filter_applied': 'all' if filter_graphs is None else ('graphs' if filter_graphs else 'non_graphs')
        }
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/collections/{collection_name}")
async def get_collection_data(
    collection_name: str,
    limit: Optional[int] = None,
    skip: Optional[int] = None,
    filter_key: Optional[str] = None
):
    """
    Query any collection in the database with optional filtering and special handling for graphs
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        collection = db.collection(collection_name)
        
        # Build AQL query based on parameters
        aql = f"FOR doc IN {collection_name}"
        
        # Add filter if specified
        if filter_key:
            aql += f" FILTER doc._key == @key"
        
        # Add limit and skip
        if skip:
            aql += f" SKIP {skip}"
        if limit:
            aql += f" LIMIT {limit}"
        
        aql += " RETURN doc"
        
        # Execute query
        cursor = db.aql.execute(
            aql,
            bind_vars={'key': filter_key} if filter_key else None
        )
        
        results = [doc for doc in cursor]
        
        # If it's a graph collection, also get vertices
        if collection_name in KNOWN_COLLECTIONS['graphs']:
            vertex_collections = set()
            for edge in results:
                vertex_collections.add(edge['_from'].split('/')[0])
                vertex_collections.add(edge['_to'].split('/')[0])
            
            vertices = []
            for vertex_col in vertex_collections:
                try:
                    if db.has_collection(vertex_col):
                        vertices.extend([v for v in db.collection(vertex_col).all()])
                except Exception as e:
                    print(f"Warning: Could not fetch vertices from {vertex_col}: {e}")
            
            return {
                'collection': collection_name,
                'type': 'graph',
                'edge_count': len(results),
                'vertex_count': len(vertices),
                'edges': results,
                'vertices': vertices
            }
        else:
            return {
                'collection': collection_name,
                'type': 'collection',
                'count': len(results),
                'data': results
            }
        
    except Exception as e:
        print(f"Error querying collection: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/collections/{collection_name}/keys")
async def get_collection_keys(collection_name: str):
    """
    Get just the _key values from a collection
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        aql = f"""
        FOR doc IN {collection_name}
        RETURN doc._key
        """
        
        cursor = db.aql.execute(aql)
        keys = [key for key in cursor]
        
        return {
            'collection': collection_name,
            'key_count': len(keys),
            'keys': keys
        }
        
    except Exception as e:
        print(f"Error getting collection keys: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/collections/{collection_name}/info")
async def get_collection_info(collection_name: str):
    """
    Get metadata about any collection
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        collection = db.collection(collection_name)
        
        return {
            "name": collection_name,
            #"type": collection_type,
            "count": collection.count(),
            "properties": collection.properties()
        }
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )