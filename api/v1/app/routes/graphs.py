from fastapi import APIRouter, HTTPException
from typing import List, Optional, Dict, Any
from arango import ArangoClient
from ..config.settings import Settings
import logging
from ..utils.path_processor import process_path_data
from ..utils.load_processor import process_load_data

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

router = APIRouter()
settings = Settings()

# Debug print to see registered routes
print("Available routes:")
for route in router.routes:
    print(f"  {route.path}")

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


###################
# Collection Routes
###################

@router.get("/graphs")
async def get_graphs(filter_graphs: Optional[bool] = None):
    """
    Get a list of graph collections in the database
    """
    try:
        db = get_db()
        # Get all collections
        collections = db.collections()
        
        # Filter out system collections (those starting with '_')
        # Then apply graph filter if specified
        graph_collections = [
            {
                'name': c['name'],
                'type': c['type'],
                'status': c['status'],
                'count': db.collection(c['name']).count()
            }
            for c in collections
            if not c['name'].startswith('_') and c['name'].endswith('_graph')
        ]
        
        # Sort by name
        graph_collections.sort(key=lambda x: x['name'])
        
        return {
            'collections': graph_collections,
            'total_count': len(graph_collections)
        }
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/graphs/{collection_name}")
async def get_graph(collection_name: str):
    """
    Get information about a specific graph collection
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Graph collection {collection_name} not found"
            )
        
        if not collection_name.endswith('_graph'):
            raise HTTPException(
                status_code=400,
                detail=f"Collection {collection_name} is not a graph collection"
            )
        
        collection = db.collection(collection_name)
        properties = collection.properties()
        
        return {
            'name': collection_name,
            'type': properties['type'],
            'status': properties['status'],
            'count': collection.count()
        }
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/graphs/{collection_name}/info")
async def get_graph_info(collection_name: str):
    """
    Get detailed information about a graph collection
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Graph collection {collection_name} not found"
            )
        
        if not collection_name.endswith('_graph'):
            raise HTTPException(
                status_code=400,
                detail=f"Collection {collection_name} is not a graph collection"
            )
        
        collection = db.collection(collection_name)
        properties = collection.properties()
        statistics = collection.statistics()
        
        # Get vertex collections connected to this graph
        vertex_collections = set()
        for edge in collection:
            vertex_collections.add(edge['_from'].split('/')[0])
            vertex_collections.add(edge['_to'].split('/')[0])
        
        return {
            'name': collection_name,
            'properties': properties,
            'statistics': statistics,
            'vertex_collections': list(vertex_collections)
        }
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

###################
# Collection Routes
###################

@router.get("/graphs/{collection_name}/vertices")
async def get_vertex_info(collection_name: str):
    """
    Get vertex information from a graph collection's edges
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        collection = db.collection(collection_name)
        
        # Debug print
        print(f"Processing vertices for collection: {collection_name}")
        
        try:
            # Get all edges to find vertex collections
            vertex_collections = set()
            vertex_info = {}
            
            # First pass: collect all vertex collections
            for edge in collection.all():
                if '_from' in edge and '_to' in edge:
                    from_collection = edge['_from'].split('/')[0]
                    to_collection = edge['_to'].split('/')[0]
                    vertex_collections.add(from_collection)
                    vertex_collections.add(to_collection)
            
            print(f"Found vertex collections: {vertex_collections}")
            
            # Second pass: get vertices from each collection
            for vertex_col in vertex_collections:
                try:
                    if db.has_collection(vertex_col):
                        vertices = []
                        for vertex in db.collection(vertex_col).all():
                            vertices.append({
                                '_id': vertex['_id'],
                                '_key': vertex['_key'],
                                'collection': vertex_col
                            })
                        vertex_info[vertex_col] = vertices
                        print(f"Processed {len(vertices)} vertices from {vertex_col}")
                except Exception as e:
                    print(f"Error processing collection {vertex_col}: {str(e)}")
                    vertex_info[vertex_col] = {"error": str(e)}
            
            return {
                'collection': collection_name,
                'vertex_collections': list(vertex_collections),
                'total_vertices': sum(len(vertices) for vertices in vertex_info.values() 
                                   if isinstance(vertices, list)),
                'vertices_by_collection': vertex_info
            }
            
        except Exception as e:
            print(f"Error processing vertices: {str(e)}")
            raise HTTPException(
                status_code=500,
                detail=f"Error processing vertices: {str(e)}"
            )
            
    except Exception as e:
        print(f"Error in get_vertex_info: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        ) 

@router.get("/graphs/{collection_name}/vertices/keys")
async def get_vertex_keys(collection_name: str):
    """
    Get just the keys of vertices referenced in a graph collection
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        collection = db.collection(collection_name)
        
        # Debug print
        print(f"Getting vertex keys for collection: {collection_name}")
        
        try:
            # Get all edges to find vertex collections
            vertex_keys = set()
            
            # First pass: collect all unique vertex keys from edges
            aql = f"""
            FOR edge IN {collection_name}
                COLLECT AGGREGATE 
                    from_keys = UNIQUE(PARSE_IDENTIFIER(edge._from).key),
                    to_keys = UNIQUE(PARSE_IDENTIFIER(edge._to).key)
                RETURN {{
                    keys: UNION_DISTINCT(from_keys, to_keys)
                }}
            """
            
            cursor = db.aql.execute(aql)
            results = [doc for doc in cursor]
            
            if results and results[0]['keys']:
                return {
                    'collection': collection_name,
                    'vertex_count': len(results[0]['keys']),
                    'vertex_keys': sorted(results[0]['keys'])
                }
            else:
                return {
                    'collection': collection_name,
                    'vertex_count': 0,
                    'vertex_keys': []
                }
            
        except Exception as e:
            print(f"Error processing vertex keys: {str(e)}")
            raise HTTPException(
                status_code=500,
                detail=f"Error processing vertex keys: {str(e)}"
            )
            
    except Exception as e:
        print(f"Error in get_vertex_keys: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        ) 

@router.get("/graphs/{collection_name}/vertices/ids")
async def get_vertex_ids(collection_name: str):
    """
    Get both _key and _id for vertices referenced in a graph collection
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Debug print
        print(f"Getting vertex IDs for collection: {collection_name}")
        
        try:
            aql = f"""
            FOR edge IN {collection_name}
                COLLECT AGGREGATE 
                    from_vertices = UNIQUE({{_id: edge._from, _key: PARSE_IDENTIFIER(edge._from).key}}),
                    to_vertices = UNIQUE({{_id: edge._to, _key: PARSE_IDENTIFIER(edge._to).key}})
                RETURN {{
                    vertices: UNION_DISTINCT(from_vertices, to_vertices)
                }}
            """
            
            cursor = db.aql.execute(aql)
            results = [doc for doc in cursor]
            
            if results and results[0]['vertices']:
                # Sort by _key for consistency
                sorted_vertices = sorted(results[0]['vertices'], key=lambda x: x['_key'])
                return {
                    'collection': collection_name,
                    'vertex_count': len(sorted_vertices),
                    'vertices': sorted_vertices
                }
            else:
                return {
                    'collection': collection_name,
                    'vertex_count': 0,
                    'vertices': []
                }
            
        except Exception as e:
            print(f"Error processing vertex IDs: {str(e)}")
            raise HTTPException(
                status_code=500,
                detail=f"Error processing vertex IDs: {str(e)}"
            )
            
    except Exception as e:
        print(f"Error in get_vertex_ids: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/graphs/{collection_name}/vertices/algo")
async def get_vertices_by_algo(collection_name: str, algo: int = 0):
    """
    Get vertices that participate in a specific Flex-Algo.
    Filters vertices based on the 'algo' field in their SRv6 endpoint behavior.
    
    Args:
        collection_name: The graph collection name
        algo: The algorithm ID to filter by (default: 0)
        
    Example:
        GET /graphs/ipv6_graph/vertices/algo?algo=129
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        print(f"Getting vertices for collection: {collection_name} with algo: {algo}")
        
        try:
            # Query to find all vertices that have the specified algo in their sids array
            aql = f"""
            FOR edge IN {collection_name}
                // Get unique vertex IDs from both ends of edges
                FOR vertex_id IN UNION_DISTINCT([edge._from], [edge._to])
                    // Parse the collection and key from the vertex ID
                    LET vertex_collection = PARSE_IDENTIFIER(vertex_id).collection
                    LET vertex_key = PARSE_IDENTIFIER(vertex_id).key
                    
                    // Fetch the actual vertex document
                    LET vertex = DOCUMENT(vertex_id)
                    
                    // Filter vertices that have sids with matching algo
                    FILTER vertex != null
                    FILTER HAS(vertex, 'sids') AND vertex.sids != null
                    FILTER LENGTH(
                        FOR sid IN vertex.sids
                            FILTER HAS(sid, 'srv6_endpoint_behavior') 
                            FILTER HAS(sid.srv6_endpoint_behavior, 'algo')
                            FILTER sid.srv6_endpoint_behavior.algo == @algo
                            RETURN sid
                    ) > 0
                    
                    // Return vertex information with SID details
                    RETURN DISTINCT {{
                        _id: vertex._id,
                        _key: vertex._key,
                        collection: vertex_collection,
                        name: HAS(vertex, 'name') ? vertex.name : null,
                        router_id: HAS(vertex, 'router_id') ? vertex.router_id : null,
                        sids: (
                            FOR sid IN vertex.sids
                                FILTER HAS(sid, 'srv6_endpoint_behavior')
                                FILTER HAS(sid.srv6_endpoint_behavior, 'algo')
                                FILTER sid.srv6_endpoint_behavior.algo == @algo
                                RETURN {{
                                    srv6_sid: sid.srv6_sid,
                                    algo: sid.srv6_endpoint_behavior.algo,
                                    endpoint_behavior: sid.srv6_endpoint_behavior.endpoint_behavior,
                                    flag: sid.srv6_endpoint_behavior.flag
                                }}
                        )
                    }}
            """
            
            cursor = db.aql.execute(aql, bind_vars={'algo': algo})
            results = [doc for doc in cursor]
            
            # Group results by collection for better organization
            vertices_by_collection = {}
            for vertex in results:
                coll = vertex['collection']
                if coll not in vertices_by_collection:
                    vertices_by_collection[coll] = []
                vertices_by_collection[coll].append(vertex)
            
            return {
                'graph_collection': collection_name,
                'algo': algo,
                'total_vertices': len(results),
                'vertex_collections': list(vertices_by_collection.keys()),
                'vertices_by_collection': vertices_by_collection
            }
            
        except Exception as e:
            print(f"Error processing vertices by algo: {str(e)}")
            raise HTTPException(
                status_code=500,
                detail=f"Error processing vertices by algo: {str(e)}"
            )
            
    except Exception as e:
        print(f"Error in get_vertices_by_algo: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        ) 

@router.get("/graphs/{collection_name}/vertices/summary")
async def get_vertex_summary(
    collection_name: str, 
    limit: int = 100,
    vertex_collection: str = None  # New optional query parameter
):
    """
    Get summarized vertex data from any graph in the database.
    Returns only key fields that have data.
    Optionally filter by specific vertex collection.
    """
    try:
        db = get_db()
        
        # First, get the vertex collections for this graph
        collections_query = """
        FOR e IN @@graph
            COLLECT AGGREGATE 
                from_cols = UNIQUE(PARSE_IDENTIFIER(e._from).collection),
                to_cols = UNIQUE(PARSE_IDENTIFIER(e._to).collection)
            RETURN {
                vertex_collections: UNION_DISTINCT(from_cols, to_cols)
            }
        """
        
        collections_cursor = db.aql.execute(
            collections_query,
            bind_vars={
                '@graph': collection_name
            }
        )
        
        collections_result = [doc for doc in collections_cursor]
        if not collections_result:
            raise HTTPException(
                status_code=404,
                detail=f"No vertex collections found for graph {collection_name}"
            )
            
        vertex_collections = collections_result[0]['vertex_collections']
        
        # If vertex_collection is specified, validate it exists in the graph
        if vertex_collection and vertex_collection not in vertex_collections:
            raise HTTPException(
                status_code=400,
                detail=f"Vertex collection '{vertex_collection}' not found in graph. Available collections: {vertex_collections}"
            )
        
        # Filter collections if vertex_collection is specified
        collections_to_query = [vertex_collection] if vertex_collection else vertex_collections
        
        # Now query each vertex collection
        all_vertices = []
        for vcoll in collections_to_query:
            vertex_query = """
            FOR v IN @@collection
                LIMIT @limit
                RETURN {
                    collection: @collection_name,
                    _key: v._key,
                    _id: v._id,
                    name: HAS(v, 'name') ? v.name : null,
                    prefix: HAS(v, 'prefix') ? v.prefix : null,
                    sids: HAS(v, 'sids') ? v.sids[*].srv6_sid : null,
                    protocol: HAS(v, 'protocol') ? v.protocol : null,
                    asn: HAS(v, 'asn') ? v.asn : null
                }
            """
            
            vertex_cursor = db.aql.execute(
                vertex_query,
                bind_vars={
                    '@collection': vcoll,
                    'collection_name': vcoll,
                    'limit': limit
                }
            )
            
            vertices = [doc for doc in vertex_cursor]
            all_vertices.extend(vertices)
        
        # Remove null fields from the response
        cleaned_vertices = []
        for vertex in all_vertices:
            cleaned_vertex = {k: v for k, v in vertex.items() if v is not None}
            cleaned_vertices.append(cleaned_vertex)
        
        return {
            'graph': collection_name,
            'vertex_collections': vertex_collections,
            'filtered_collection': vertex_collection,
            'total_vertices': len(cleaned_vertices),
            'vertices': cleaned_vertices
        }
        
    except Exception as e:
        print(f"Error getting vertex summary: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

################
# Edge Routes
################

@router.get("/graphs/{collection_name}/edges")
async def get_edge_connections(collection_name: str):
    """
    Get only the _from and _to fields from a graph collection
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        collection = db.collection(collection_name)
        
        # Debug print
        print(f"Collection properties: {collection.properties()}")
        
        # Get all edges with error handling
        try:
            edges = []
            cursor = collection.all()
            for edge in cursor:
                if '_from' in edge and '_to' in edge:
                    edges.append({
                        '_from': edge['_from'],
                        '_to': edge['_to']
                    })
                else:
                    print(f"Warning: Edge missing _from or _to: {edge}")
            
            print(f"Found {len(edges)} edges")
            
            return {
                'collection': collection_name,
                'edge_count': len(edges),
                'edges': edges
            }
            
        except Exception as e:
            print(f"Error processing edges: {str(e)}")
            raise HTTPException(
                status_code=500,
                detail=f"Error processing edges: {str(e)}"
            )
            
    except Exception as e:
        print(f"Error in get_edge_connections: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        ) 

@router.get("/graphs/{collection_name}/edges/detail")
async def get_detailed_edge_connections(collection_name: str, limit: Optional[int] = None):
    """
    Get detailed edge information from a graph collection including metrics and properties
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        collection = db.collection(collection_name)
        
        # Get edges with additional fields
        try:
            edges = []
            cursor = collection.all()
            for edge in cursor:
                if '_from' in edge and '_to' in edge:
                    edge_detail = {
                        '_key': edge.get('_key'),
                        '_from': edge['_from'],
                        '_to': edge['_to'],
                        'name': edge.get('name'),
                        'prefix': edge.get('prefix'),
                        'protocol': edge.get('protocol'),
                        'sids': edge.get('sids', []),
                        'country_codes': edge.get('country_codes'),
                        'metrics': {
                            'unidir_delay': edge.get('unidir_link_delay'),
                            'percent_util_out': edge.get('percent_util_out'),
                            'percent_util_in': edge.get('percent_util_in'),
                            'bandwidth': edge.get('max_link_bandwidth'),
                            'reservable_bandwidth': edge.get('max_reservable_link_bandwidth'),
                            'load': edge.get('load')
                        },
                        'timestamps': {
                            'first_seen': edge.get('first_seen_at'),
                            'last_seen': edge.get('last_seen_at'),
                            'updated': edge.get('updated_at')
                        }
                    }
                    
                    # Remove any metrics that are None
                    edge_detail['metrics'] = {k: v for k, v in edge_detail['metrics'].items() if v is not None}
                    edge_detail['timestamps'] = {k: v for k, v in edge_detail['timestamps'].items() if v is not None}
                    
                    # Only include non-None fields
                    edges.append({k: v for k, v in edge_detail.items() if v is not None})
                else:
                    print(f"Warning: Edge missing _from or _to: {edge}")
            
            # Apply limit if specified
            result_edges = edges[:limit] if limit else edges
            
            return {
                'collection': collection_name,
                'edge_count': len(edges),
                'returned_edges': len(result_edges),
                'edges': result_edges
            }
            
        except Exception as e:
            print(f"Error processing edges: {str(e)}")
            raise HTTPException(
                status_code=500,
                detail=f"Error processing edges: {str(e)}"
            )
            
    except Exception as e:
        print(f"Error in get_detailed_edge_connections: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

################
# Topology Route
################

@router.get("/graphs/{collection_name}/topology")
async def get_topology(
    collection_name: str, 
    include_all_fields: bool = True  # New optional parameter
):
    """
    Get complete topology information with optional field filtering
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Get all edges
        collection = db.collection(collection_name)
        edges = []
        vertex_ids = set()
        
        # Get all edges with full data
        cursor = collection.all()
        for edge in cursor:
            if '_from' in edge and '_to' in edge:
                # Include all fields if requested
                if include_all_fields:
                    edges.append(edge)
                else:
                    # Include only basic fields
                    edges.append({
                        '_from': edge['_from'],
                        '_to': edge['_to']
                    })
                vertex_ids.add(edge['_from'])
                vertex_ids.add(edge['_to'])
        
        # Get vertex details
        vertices = {}
        for vertex_id in vertex_ids:
            collection_name, key = vertex_id.split('/')
            
            try:
                vertex = db.collection(collection_name).get(key)
                if vertex:
                    if include_all_fields:
                        # Include all vertex fields
                        vertices[vertex_id] = vertex
                    else:
                        # Include only commonly used fields
                        vertex_detail = {
                            'collection': collection_name,
                            'name': vertex.get('name'),
                            'prefix': vertex.get('prefix'),
                            'protocol': vertex.get('protocol'),
                            'sids': [sid.get('srv6_sid') for sid in vertex.get('sids', []) if 'srv6_sid' in sid],
                            'asn': vertex.get('asn')
                        }
                        # Remove None values
                        vertices[vertex_id] = {k: v for k, v in vertex_detail.items() if v is not None}
            except Exception as vertex_error:
                print(f"Error getting vertex {vertex_id}: {str(vertex_error)}")
                continue
        
        return {
            'edges': edges,
            'vertices': vertices,
            'total_edges': len(edges),
            'total_vertices': len(vertices),
            'include_all_fields': include_all_fields
        }
            
    except Exception as e:
        print(f"Error in get_topology: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/graphs/{collection_name}/topology/nodes")
async def get_node_topology(
    collection_name: str, 
    include_all_fields: bool = True  # Default to returning all fields
):
    """
    Get topology information filtered to only node-to-node connections
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Get edges filtered for node connections
        edge_query = """
        FOR edge IN @@collection
            FILTER CONTAINS(edge._from, 'node') AND CONTAINS(edge._to, 'node')
            RETURN edge
        """
        
        edge_cursor = db.aql.execute(
            edge_query,
            bind_vars={
                '@collection': collection_name
            }
        )
        
        # Process edges based on include_all_fields
        edges = []
        vertex_ids = set()
        
        for edge in edge_cursor:
            if '_from' in edge and '_to' in edge:
                if include_all_fields:
                    edges.append(edge)
                else:
                    edges.append({
                        '_from': edge['_from'],
                        '_to': edge['_to']
                    })
                vertex_ids.add(edge['_from'])
                vertex_ids.add(edge['_to'])
        
        # Get vertex details
        vertices = {}
        for vertex_id in vertex_ids:
            collection_name, key = vertex_id.split('/')
            
            try:
                vertex = db.collection(collection_name).get(key)
                if vertex:
                    if include_all_fields:
                        # Include all vertex fields
                        vertices[vertex_id] = vertex
                    else:
                        # Include only commonly used fields
                        vertex_detail = {
                            'collection': collection_name,
                            'name': vertex.get('name'),
                            'prefix': vertex.get('prefix'),
                            'protocol': vertex.get('protocol'),
                            'sids': [sid.get('srv6_sid') for sid in vertex.get('sids', []) if 'srv6_sid' in sid],
                            'asn': vertex.get('asn')
                        }
                        # Remove None values
                        vertices[vertex_id] = {k: v for k, v in vertex_detail.items() if v is not None}
            except Exception as vertex_error:
                print(f"Error getting vertex {vertex_id}: {str(vertex_error)}")
                continue
        
        return {
            'edges': edges,
            'vertices': vertices,
            'total_edges': len(edges),
            'total_vertices': len(vertices),
            'include_all_fields': include_all_fields
        }
            
    except Exception as e:
        print(f"Error in get_node_topology: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/graphs/{collection_name}/topology/nodes/algo")
async def get_node_topology_by_algo(
    collection_name: str,
    algo: int = 0,
    include_all_fields: bool = True
):
    """
    Get topology information filtered to only node-to-node connections
    that participate in a specific Flex-Algo.
    
    Args:
        collection_name: The graph collection name
        algo: The algorithm ID to filter by (default: 0)
        include_all_fields: Return all fields or just essential ones (default: True)
        
    Example:
        GET /graphs/ipv6_graph/topology/nodes/algo?algo=128
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        print(f"Getting topology for collection: {collection_name} with algo: {algo}")
        
        # First, get all vertices that participate in this algo
        vertices_with_algo = set()
        
        # Query to find vertices with the specified algo
        vertex_query = f"""
        FOR edge IN {collection_name}
            FILTER CONTAINS(edge._from, 'node') AND CONTAINS(edge._to, 'node')
            FOR vertex_id IN UNION_DISTINCT([edge._from], [edge._to])
                LET vertex = DOCUMENT(vertex_id)
                FILTER vertex != null
                FILTER HAS(vertex, 'sids') AND vertex.sids != null
                FILTER LENGTH(
                    FOR sid IN vertex.sids
                        FILTER HAS(sid, 'srv6_endpoint_behavior')
                        FILTER HAS(sid.srv6_endpoint_behavior, 'algo')
                        FILTER sid.srv6_endpoint_behavior.algo == @algo
                        RETURN sid
                ) > 0
                RETURN DISTINCT vertex_id
        """
        
        vertex_cursor = db.aql.execute(vertex_query, bind_vars={'algo': algo})
        vertices_with_algo = set([vid for vid in vertex_cursor])
        
        print(f"Found {len(vertices_with_algo)} vertices with algo {algo}")
        
        # Now get edges where BOTH endpoints participate in this algo
        edge_query = f"""
        FOR edge IN {collection_name}
            FILTER CONTAINS(edge._from, 'node') AND CONTAINS(edge._to, 'node')
            RETURN edge
        """
        
        edge_cursor = db.aql.execute(edge_query)
        
        # Filter edges where both endpoints have the algo
        edges = []
        filtered_vertex_ids = set()
        
        for edge in edge_cursor:
            if '_from' in edge and '_to' in edge:
                # Only include edge if both vertices support this algo
                if edge['_from'] in vertices_with_algo and edge['_to'] in vertices_with_algo:
                    if include_all_fields:
                        edges.append(edge)
                    else:
                        edges.append({
                            '_from': edge['_from'],
                            '_to': edge['_to']
                        })
                    filtered_vertex_ids.add(edge['_from'])
                    filtered_vertex_ids.add(edge['_to'])
        
        # Get vertex details for vertices that are actually used in edges
        vertices = {}
        for vertex_id in filtered_vertex_ids:
            vertex_collection, key = vertex_id.split('/')
            
            try:
                vertex = db.collection(vertex_collection).get(key)
                if vertex:
                    if include_all_fields:
                        # Include all vertex fields
                        vertices[vertex_id] = vertex
                    else:
                        # Filter SIDs to only include those matching the algo
                        algo_sids = []
                        if 'sids' in vertex and vertex['sids']:
                            for sid in vertex['sids']:
                                if ('srv6_endpoint_behavior' in sid and 
                                    'algo' in sid['srv6_endpoint_behavior'] and
                                    sid['srv6_endpoint_behavior']['algo'] == algo):
                                    algo_sids.append({
                                        'srv6_sid': sid.get('srv6_sid'),
                                        'algo': sid['srv6_endpoint_behavior'].get('algo'),
                                        'endpoint_behavior': sid['srv6_endpoint_behavior'].get('endpoint_behavior')
                                    })
                        
                        vertex_detail = {
                            'collection': vertex_collection,
                            'name': vertex.get('name'),
                            'router_id': vertex.get('router_id'),
                            'sids': algo_sids if algo_sids else None,
                            'asn': vertex.get('asn')
                        }
                        # Remove None values
                        vertices[vertex_id] = {k: v for k, v in vertex_detail.items() if v is not None}
            except Exception as vertex_error:
                print(f"Error getting vertex {vertex_id}: {str(vertex_error)}")
                continue
        
        return {
            'graph_collection': collection_name,
            'algo': algo,
            'edges': edges,
            'vertices': vertices,
            'total_edges': len(edges),
            'total_vertices': len(vertices),
            'include_all_fields': include_all_fields
        }

    except Exception as e:
        print(f"Error in get_node_topology_by_algo: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

##############################
# Shortest Path and Traversals
##############################

# basic shortest path
@router.get("/graphs/{collection_name}/shortest_path")
async def get_shortest_path(
    collection_name: str,
    source: str,
    destination: str,
    direction: str = "outbound",  # or "inbound", "any"
    algo: int = 0  # Flex-Algo to use for SRv6 SID selection
):
    """
    Find shortest path between two nodes in a graph with detailed vertex and edge information.
    
    Args:
        collection_name: The graph collection to search
        source: Source node ID
        destination: Destination node ID
        direction: Path direction (outbound, inbound, or any)
        algo: Flex-Algo ID for SRv6 SID selection (default: 0)
    
    The algo parameter filters which SRv6 SIDs are used in the srv6_data response.
    Only SIDs matching the specified algo will be included in the SRv6 USID calculation.
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Validate direction parameter
        if direction.lower() not in ["outbound", "inbound", "any"]:
            raise HTTPException(
                status_code=400,
                detail="Direction must be 'outbound', 'inbound', or 'any'"
            )
        
        # Build AQL query with optional algo filtering for igp_nodes
        # For algo 0 or when algo filtering is not needed, use standard shortest path
        # For non-zero algo, filter igp_nodes to only include those participating in that algo
        if algo == 0:
            # Standard shortest path without algo filtering
            aql = f"""
            WITH igp_node
            LET path = (
                FOR v, e IN {direction.upper()}
                    SHORTEST_PATH @source TO @destination
                    @graph_name
                    RETURN {{
                        vertex: {{
                            _id: v._id,
                            _key: v._key,
                            router_id: v.router_id,
                            prefix: v.prefix,
                            name: v.name,
                            sids: v.sids
                        }},
                        edge: e ? {{
                            _id: e._id,
                            _key: e._key,
                            _from: e._from,
                            _to: e._to,
                            latency: e.latency,
                            percent_util_out: e.percent_util_out,
                            load: e.load
                        }} : null
                    }}
            )
            RETURN {{
                path: path,
                hopcount: LENGTH(path) - 1,
                vertex_count: LENGTH(path),
                source_info: FIRST(path).vertex,
                destination_info: LAST(path).vertex
            }}
            """
        else:
            # Algo-aware shortest path - use K_SHORTEST_PATHS to find multiple paths
            # and filter to get the first one where all igp_nodes support the algo
            aql = f"""
            WITH igp_node
            FOR path IN {direction.upper()}
                K_SHORTEST_PATHS @source TO @destination
                @graph_name
                // Check if all igp_nodes in this path support the requested algo
                LET igp_nodes_in_path = (
                    FOR v IN path.vertices
                        FILTER CONTAINS(v._id, 'igp_node')
                        RETURN v
                )
                LET nodes_with_algo = (
                    FOR node IN igp_nodes_in_path
                        FILTER HAS(node, 'sids') AND node.sids != null
                        FILTER LENGTH(
                            FOR sid IN node.sids
                                FILTER HAS(sid, 'srv6_endpoint_behavior')
                                FILTER HAS(sid.srv6_endpoint_behavior, 'algo')
                                FILTER sid.srv6_endpoint_behavior.algo == @algo
                                RETURN sid
                        ) > 0
                        RETURN node
                )
                // Only accept paths where all igp_nodes support the algo
                FILTER LENGTH(igp_nodes_in_path) == LENGTH(nodes_with_algo)
                LIMIT 1
                
                LET formatted_path = (
                    FOR i IN 0..LENGTH(path.vertices)-1
                        RETURN {{
                            vertex: {{
                                _id: path.vertices[i]._id,
                                _key: path.vertices[i]._key,
                                router_id: path.vertices[i].router_id,
                                prefix: path.vertices[i].prefix,
                                name: path.vertices[i].name,
                                sids: path.vertices[i].sids
                            }},
                            edge: i < LENGTH(path.edges) ? {{
                                _id: path.edges[i]._id,
                                _key: path.edges[i]._key,
                                _from: path.edges[i]._from,
                                _to: path.edges[i]._to,
                                latency: path.edges[i].latency,
                                percent_util_out: path.edges[i].percent_util_out,
                                load: path.edges[i].load
                            }} : null
                        }}
                )
                
                RETURN {{
                    path: formatted_path,
                    hopcount: LENGTH(path.vertices) - 1,
                    vertex_count: LENGTH(path.vertices),
                    source_info: {{
                        _id: path.vertices[0]._id,
                        _key: path.vertices[0]._key,
                        router_id: path.vertices[0].router_id,
                        prefix: path.vertices[0].prefix,
                        name: path.vertices[0].name,
                        sids: path.vertices[0].sids
                    }},
                    destination_info: {{
                        _id: LAST(path.vertices)._id,
                        _key: LAST(path.vertices)._key,
                        router_id: LAST(path.vertices).router_id,
                        prefix: LAST(path.vertices).prefix,
                        name: LAST(path.vertices).name,
                        sids: LAST(path.vertices).sids
                    }}
                }}
            """
        
        # Prepare bind variables
        bind_vars = {
            'source': source,
            'destination': destination,
            'graph_name': collection_name
        }
        
        # Add algo to bind vars if filtering is enabled
        if algo != 0:
            bind_vars['algo'] = algo
        
        cursor = db.aql.execute(aql, bind_vars=bind_vars)
        
        results = [doc for doc in cursor]
        
        if not results or not results[0]['path']:
            return {
                "found": False,
                "message": "No path found between specified nodes"
            }
        
        # Get the existing response
        response = {
            "found": True,
            "path": results[0]['path'],
            "hopcount": results[0]['hopcount'],
            "vertex_count": results[0]['vertex_count'],
            "source_info": results[0]['source_info'],
            "destination_info": results[0]['destination_info'],
            "direction": direction,
            "algo": algo
        }
        
        # Process and append the SRv6 data with algo filtering
        srv6_data = process_path_data(results[0]['path'], source, destination, algo=algo)
        response["srv6_data"] = srv6_data
        
        return response
        
    except Exception as e:
        print(f"Error finding shortest path: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

# latency weighted shortest path
@router.get("/graphs/{collection_name}/shortest_path/latency")
async def get_shortest_path_latency(
    collection_name: str,
    source: str,
    destination: str,
    direction: str = "outbound",  # or "inbound", "any"
    algo: int = 0  # Flex-Algo to use for SRv6 SID selection
):
    """
    Find shortest path between two nodes using latency as weight.
    
    Args:
        collection_name: The graph collection to search
        source: Source node ID
        destination: Destination node ID
        direction: Path direction (outbound, inbound, or any)
        algo: Flex-Algo ID for path computation and SRv6 SID selection (default: 0)
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Validate direction parameter
        if direction.lower() not in ["outbound", "inbound", "any"]:
            raise HTTPException(
                status_code=400,
                detail="Direction must be 'outbound', 'inbound', or 'any'"
            )
        
        # Build AQL query with optional algo filtering
        if algo == 0:
            # Standard shortest path with latency weight
            aql = f"""
            WITH igp_node
            LET path = (
                FOR v, e IN {direction.upper()}
                    SHORTEST_PATH @source TO @destination
                    @graph_name
                    OPTIONS {{
                        weightAttribute: 'latency',
                        defaultWeight: 1
                    }}
                    RETURN {{
                        vertex: {{
                            _id: v._id,
                            _key: v._key,
                            router_id: v.router_id,
                            prefix: v.prefix,
                            name: v.name,
                            sids: v.sids
                        }},
                        edge: e ? {{
                            _id: e._id,
                            _key: e._key,
                            _from: e._from,
                            _to: e._to,
                            latency: e.latency,
                            percent_util_out: e.percent_util_out,
                            load: e.load
                        }} : null
                    }}
            )
            
            LET total_latency = (
                FOR p IN path
                    FILTER p.edge != null
                    COLLECT AGGREGATE total = SUM(p.edge.latency)
                    RETURN total
            )
            
            RETURN {{
                path: path,
                hopcount: LENGTH(path) - 1,
                vertex_count: LENGTH(path),
                source_info: FIRST(path).vertex,
                destination_info: LAST(path).vertex,
                total_latency: FIRST(total_latency)
            }}
            """
        else:
            # Algo-aware shortest path with latency weight
            aql = f"""
            WITH igp_node
            FOR path IN {direction.upper()}
                K_SHORTEST_PATHS @source TO @destination
                @graph_name
                OPTIONS {{
                    weightAttribute: 'latency',
                    defaultWeight: 1
                }}
                LET igp_nodes_in_path = (
                    FOR v IN path.vertices
                        FILTER CONTAINS(v._id, 'igp_node')
                        RETURN v
                )
                LET nodes_with_algo = (
                    FOR node IN igp_nodes_in_path
                        FILTER HAS(node, 'sids') AND node.sids != null
                        FILTER LENGTH(
                            FOR sid IN node.sids
                                FILTER HAS(sid, 'srv6_endpoint_behavior')
                                FILTER HAS(sid.srv6_endpoint_behavior, 'algo')
                                FILTER sid.srv6_endpoint_behavior.algo == @algo
                                RETURN sid
                        ) > 0
                        RETURN node
                )
                FILTER LENGTH(igp_nodes_in_path) == LENGTH(nodes_with_algo)
                LIMIT 1
                
                LET formatted_path = (
                    FOR i IN 0..LENGTH(path.vertices)-1
                        RETURN {{
                            vertex: {{
                                _id: path.vertices[i]._id,
                                _key: path.vertices[i]._key,
                                router_id: path.vertices[i].router_id,
                                prefix: path.vertices[i].prefix,
                                name: path.vertices[i].name,
                                sids: path.vertices[i].sids
                            }},
                            edge: i < LENGTH(path.edges) ? {{
                                _id: path.edges[i]._id,
                                _key: path.edges[i]._key,
                                _from: path.edges[i]._from,
                                _to: path.edges[i]._to,
                                latency: path.edges[i].latency,
                                percent_util_out: path.edges[i].percent_util_out,
                                load: path.edges[i].load
                            }} : null
                        }}
                )
                
                LET total_latency = (
                    FOR i IN 0..LENGTH(path.edges)-1
                        FILTER path.edges[i].latency != null
                        COLLECT AGGREGATE total = SUM(path.edges[i].latency)
                        RETURN total
                )
                
                RETURN {{
                    path: formatted_path,
                    hopcount: LENGTH(path.vertices) - 1,
                    vertex_count: LENGTH(path.vertices),
                    source_info: {{
                        _id: path.vertices[0]._id,
                        _key: path.vertices[0]._key,
                        router_id: path.vertices[0].router_id,
                        prefix: path.vertices[0].prefix,
                        name: path.vertices[0].name,
                        sids: path.vertices[0].sids
                    }},
                    destination_info: {{
                        _id: LAST(path.vertices)._id,
                        _key: LAST(path.vertices)._key,
                        router_id: LAST(path.vertices).router_id,
                        prefix: LAST(path.vertices).prefix,
                        name: LAST(path.vertices).name,
                        sids: LAST(path.vertices).sids
                    }},
                    total_latency: FIRST(total_latency)
                }}
            """
        
        # Prepare bind variables
        bind_vars = {
            'source': source,
            'destination': destination,
            'graph_name': collection_name
        }
        
        if algo != 0:
            bind_vars['algo'] = algo
        
        cursor = db.aql.execute(aql, bind_vars=bind_vars)
        
        results = [doc for doc in cursor]
        
        if not results or not results[0]['path']:
            return {
                "found": False,
                "message": "No path found between specified nodes"
            }
        
        # Get the existing response
        response = {
            "found": True,
            "path": results[0]['path'],
            "hopcount": results[0]['hopcount'],
            "vertex_count": results[0]['vertex_count'],
            "source_info": results[0]['source_info'],
            "destination_info": results[0]['destination_info'],
            "direction": direction,
            "total_latency": results[0]['total_latency'],
            "algo": algo
        }
        
        # Process and append the SRv6 data with algo filtering
        srv6_data = process_path_data(results[0]['path'], source, destination, algo=algo)
        response["srv6_data"] = srv6_data
        
        return response
        
    except Exception as e:
        print(f"Error finding shortest path with latency weight: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

# weighted shortest path - outbound utilization
@router.get("/graphs/{collection_name}/shortest_path/utilization")
async def get_shortest_path_utilization(
    collection_name: str,
    source: str,
    destination: str,
    direction: str = "outbound",  # or "inbound", "any"
    algo: int = 0  # Flex-Algo to use for SRv6 SID selection
):
    """
    Find shortest path between two nodes using utilization as weight.
    
    Args:
        algo: Flex-Algo ID for path computation and SRv6 SID selection (default: 0)
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Validate direction parameter
        if direction.lower() not in ["outbound", "inbound", "any"]:
            raise HTTPException(
                status_code=400,
                detail="Direction must be 'outbound', 'inbound', or 'any'"
            )
        
        # Build AQL query with optional algo filtering
        if algo == 0:
            # Standard shortest path with utilization weight
            aql = f"""
            WITH igp_node
            LET path = (
                FOR v, e IN {direction.upper()}
                    SHORTEST_PATH @source TO @destination
                    @graph_name
                    OPTIONS {{
                        weightAttribute: 'percent_util_out',
                        defaultWeight: 1
                    }}
                    RETURN {{
                        vertex: {{
                            _id: v._id,
                            _key: v._key,
                            router_id: v.router_id,
                            prefix: v.prefix,
                            name: v.name,
                            sids: v.sids
                        }},
                        edge: e ? {{
                            _id: e._id,
                            _key: e._key,
                            _from: e._from,
                            _to: e._to,
                            latency: e.latency,
                            percent_util_out: e.percent_util_out,
                            load: e.load
                        }} : null
                    }}
            )
            
            LET avg_utilization = (
                FOR p IN path
                    FILTER p.edge != null
                    COLLECT AGGREGATE 
                        avg = AVERAGE(p.edge.percent_util_out)
                    RETURN avg
            )
            
            RETURN {{
                path: path,
                hopcount: LENGTH(path) - 1,
                vertex_count: LENGTH(path),
                source_info: FIRST(path).vertex,
                destination_info: LAST(path).vertex,
                average_utilization: FIRST(avg_utilization)
            }}
            """
        else:
            # Algo-aware shortest path with utilization weight
            aql = f"""
            WITH igp_node
            FOR path IN {direction.upper()}
                K_SHORTEST_PATHS @source TO @destination
                @graph_name
                OPTIONS {{
                    weightAttribute: 'percent_util_out',
                    defaultWeight: 1
                }}
                LET igp_nodes_in_path = (
                    FOR v IN path.vertices
                        FILTER CONTAINS(v._id, 'igp_node')
                        RETURN v
                )
                LET nodes_with_algo = (
                    FOR node IN igp_nodes_in_path
                        FILTER HAS(node, 'sids') AND node.sids != null
                        FILTER LENGTH(
                            FOR sid IN node.sids
                                FILTER HAS(sid, 'srv6_endpoint_behavior')
                                FILTER HAS(sid.srv6_endpoint_behavior, 'algo')
                                FILTER sid.srv6_endpoint_behavior.algo == @algo
                                RETURN sid
                        ) > 0
                        RETURN node
                )
                FILTER LENGTH(igp_nodes_in_path) == LENGTH(nodes_with_algo)
                LIMIT 1
                
                LET formatted_path = (
                    FOR i IN 0..LENGTH(path.vertices)-1
                        RETURN {{
                            vertex: {{
                                _id: path.vertices[i]._id,
                                _key: path.vertices[i]._key,
                                router_id: path.vertices[i].router_id,
                                prefix: path.vertices[i].prefix,
                                name: path.vertices[i].name,
                                sids: path.vertices[i].sids
                            }},
                            edge: i < LENGTH(path.edges) ? {{
                                _id: path.edges[i]._id,
                                _key: path.edges[i]._key,
                                _from: path.edges[i]._from,
                                _to: path.edges[i]._to,
                                latency: path.edges[i].latency,
                                percent_util_out: path.edges[i].percent_util_out,
                                load: path.edges[i].load
                            }} : null
                        }}
                )
                
                LET avg_utilization = (
                    FOR i IN 0..LENGTH(path.edges)-1
                        FILTER path.edges[i].percent_util_out != null
                        COLLECT AGGREGATE avg = AVERAGE(path.edges[i].percent_util_out)
                        RETURN avg
                )
                
                RETURN {{
                    path: formatted_path,
                    hopcount: LENGTH(path.vertices) - 1,
                    vertex_count: LENGTH(path.vertices),
                    source_info: {{
                        _id: path.vertices[0]._id,
                        _key: path.vertices[0]._key,
                        router_id: path.vertices[0].router_id,
                        prefix: path.vertices[0].prefix,
                        name: path.vertices[0].name,
                        sids: path.vertices[0].sids
                    }},
                    destination_info: {{
                        _id: LAST(path.vertices)._id,
                        _key: LAST(path.vertices)._key,
                        router_id: LAST(path.vertices).router_id,
                        prefix: LAST(path.vertices).prefix,
                        name: LAST(path.vertices).name,
                        sids: LAST(path.vertices).sids
                    }},
                    average_utilization: FIRST(avg_utilization)
                }}
            """
        
        # Prepare bind variables
        bind_vars = {
            'source': source,
            'destination': destination,
            'graph_name': collection_name
        }
        
        if algo != 0:
            bind_vars['algo'] = algo
        
        cursor = db.aql.execute(aql, bind_vars=bind_vars)
        
        results = [doc for doc in cursor]
        
        if not results or not results[0]['path']:
            return {
                "found": False,
                "message": "No path found between specified nodes"
            }
        
        # Get the existing response
        response = {
            "found": True,
            "path": results[0]['path'],
            "hopcount": results[0]['hopcount'],
            "vertex_count": results[0]['vertex_count'],
            "source_info": results[0]['source_info'],
            "destination_info": results[0]['destination_info'],
            "direction": direction,
            "average_utilization": results[0]['average_utilization']
        }
        
        # Process and append the SRv6 data with algo filtering
        srv6_data = process_path_data(results[0]['path'], source, destination, algo=algo)
        response["srv6_data"] = srv6_data
        response["algo"] = algo
        
        return response
        
    except Exception as e:
        print(f"Error finding shortest path with utilization weight: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/graphs/{collection_name}/shortest_path/sovereignty")
async def get_shortest_path_sovereignty(
    collection_name: str,
    source: str,
    destination: str,
    excluded_countries: str,
    direction: str = "outbound",
    algo: int = 0  # Flex-Algo to use for SRv6 SID selection
):
    """
    Find shortest path between two nodes while avoiding specified countries.
    
    Args:
        algo: Flex-Algo ID for path computation and SRv6 SID selection (default: 0)
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Convert comma-separated countries to list and create filter conditions
        countries = [c.strip().upper() for c in excluded_countries.split(',')]
        country_filters = ' AND '.join([f'p.edges[*].country_codes !like "%{country}%"' for country in countries])
        
        # Build algo filtering if needed
        if algo == 0:
            algo_filter = ""
        else:
            algo_filter = f"""
            LET igp_nodes_in_path = (
                FOR v IN p.vertices
                    FILTER CONTAINS(v._id, 'igp_node')
                    RETURN v
            )
            LET nodes_with_algo = (
                FOR node IN igp_nodes_in_path
                    FILTER HAS(node, 'sids') AND node.sids != null
                    FILTER LENGTH(
                        FOR sid IN node.sids
                            FILTER HAS(sid, 'srv6_endpoint_behavior')
                            FILTER HAS(sid.srv6_endpoint_behavior, 'algo')
                            FILTER sid.srv6_endpoint_behavior.algo == {algo}
                            RETURN sid
                    ) > 0
                    RETURN node
            )
            FILTER LENGTH(igp_nodes_in_path) == LENGTH(nodes_with_algo)
            """
        
        # AQL query matching the working manual query but with additional path details
        aql = f"""
        FOR p IN {direction.upper()} k_shortest_paths
            '{source}' TO '{destination}'
            {collection_name}
            OPTIONS {{uniqueVertices: "path", bfs: true}}
            FILTER {country_filters}
            {algo_filter}
            LIMIT 1
            RETURN {{
                path: (
                    FOR v IN p.vertices
                    RETURN {{
                        vertex: {{
                            _id: v._id,
                            _key: v._key,
                            name: v.name,
                            sids: v.sids
                        }}
                    }}
                ),
                countries_traversed: p.edges[*].country_codes[*],
                hopcount: LENGTH(p.vertices) - 1,
                vertex_count: LENGTH(p.vertices),
                source_info: FIRST(p.vertices),
                destination_info: LAST(p.vertices)
            }}
        """
        
        cursor = db.aql.execute(aql)
        results = [doc for doc in cursor]
        
        if not results:
            return {
                "found": False,
                "message": f"No path found between specified nodes avoiding countries: {excluded_countries}"
            }
        
        # Create response with summary data
        response = {
            "found": True,
            "path": results[0]['path'],
            "hopcount": results[0]['hopcount'],
            "vertex_count": results[0]['vertex_count'],
            "source_info": results[0]['source_info'],
            "destination_info": results[0]['destination_info'],
            "direction": direction,
            "countries_traversed": results[0]['countries_traversed'],
            "excluded_countries": countries,
            "algo": algo
        }
        
        # Process and append the SRv6 data with algo filtering
        srv6_data = process_path_data(results[0]['path'], source, destination, algo=algo)
        response["srv6_data"] = srv6_data
        
        return response
        
    except Exception as e:
        print(f"Error finding path with sovereignty constraints: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

# weighted shortest path - load
@router.get("/graphs/{collection_name}/shortest_path/load")
async def get_shortest_path_load(
    collection_name: str,
    source: str,
    destination: str,
    direction: str = "outbound",  # or "inbound", "any"
    algo: int = 0  # Flex-Algo to use for SRv6 SID selection
):
    """
    Find shortest path between two nodes using load as weight.
    
    Args:
        algo: Flex-Algo ID for path computation and SRv6 SID selection (default: 0)
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Validate direction parameter
        if direction.lower() not in ["outbound", "inbound", "any"]:
            raise HTTPException(
                status_code=400,
                detail="Direction must be 'outbound', 'inbound', or 'any'"
            )
        
        # Build AQL query with optional algo filtering
        if algo == 0:
            # Standard shortest path with load weight
            aql = f"""
            WITH igp_node
            LET path = (
                FOR v, e IN {direction.upper()}
                    SHORTEST_PATH @source TO @destination
                    @graph_name
                    OPTIONS {{
                        weightAttribute: 'load',
                        defaultWeight: 1
                    }}
                    RETURN {{
                        vertex: {{
                            _id: v._id,
                            _key: v._key,
                            router_id: v.router_id,
                            ipv4_address: v.ipv4_address,
                            ipv6_address: v.ipv6_address,
                            prefix: v.prefix,
                            prefix_len: v.prefix_len,
                            name: v.name,
                            sids: v.sids
                        }},
                        edge: e ? {{
                            _id: e._id,
                            _key: e._key,
                            _from: e._from,
                            _to: e._to,
                            latency: e.latency,
                            percent_util_out: e.percent_util_out,
                            load: e.load
                        }} : null
                    }}
            )
            
            LET avg_load = (
                FOR p IN path
                    FILTER p.edge != null
                    COLLECT AGGREGATE 
                        avg = AVERAGE(p.edge.load)
                    RETURN avg
            )
            
            RETURN {{
                path: path,
                hopcount: LENGTH(path) - 1,
                vertex_count: LENGTH(path),
                source_info: FIRST(path).vertex,
                destination_info: LAST(path).vertex,
                average_load: FIRST(avg_load)
            }}
            """
        else:
            # Algo-aware shortest path with load weight
            aql = f"""
            WITH igp_node
            FOR path IN {direction.upper()}
                K_SHORTEST_PATHS @source TO @destination
                @graph_name
                OPTIONS {{
                    weightAttribute: 'load',
                    defaultWeight: 1
                }}
                LET igp_nodes_in_path = (
                    FOR v IN path.vertices
                        FILTER CONTAINS(v._id, 'igp_node')
                        RETURN v
                )
                LET nodes_with_algo = (
                    FOR node IN igp_nodes_in_path
                        FILTER HAS(node, 'sids') AND node.sids != null
                        FILTER LENGTH(
                            FOR sid IN node.sids
                                FILTER HAS(sid, 'srv6_endpoint_behavior')
                                FILTER HAS(sid.srv6_endpoint_behavior, 'algo')
                                FILTER sid.srv6_endpoint_behavior.algo == @algo
                                RETURN sid
                        ) > 0
                        RETURN node
                )
                FILTER LENGTH(igp_nodes_in_path) == LENGTH(nodes_with_algo)
                LIMIT 1
                
                LET formatted_path = (
                    FOR i IN 0..LENGTH(path.vertices)-1
                        RETURN {{
                            vertex: {{
                                _id: path.vertices[i]._id,
                                _key: path.vertices[i]._key,
                                router_id: path.vertices[i].router_id,
                                ipv4_address: path.vertices[i].ipv4_address,
                                ipv6_address: path.vertices[i].ipv6_address,
                                prefix: path.vertices[i].prefix,
                                prefix_len: path.vertices[i].prefix_len,
                                name: path.vertices[i].name,
                                sids: path.vertices[i].sids
                            }},
                            edge: i < LENGTH(path.edges) ? {{
                                _id: path.edges[i]._id,
                                _key: path.edges[i]._key,
                                _from: path.edges[i]._from,
                                _to: path.edges[i]._to,
                                latency: path.edges[i].latency,
                                percent_util_out: path.edges[i].percent_util_out,
                                load: path.edges[i].load
                            }} : null
                        }}
                )
                
                LET avg_load = (
                    FOR i IN 0..LENGTH(path.edges)-1
                        FILTER path.edges[i].load != null
                        COLLECT AGGREGATE avg = AVERAGE(path.edges[i].load)
                        RETURN avg
                )
                
                RETURN {{
                    path: formatted_path,
                    hopcount: LENGTH(path.vertices) - 1,
                    vertex_count: LENGTH(path.vertices),
                    source_info: {{
                        _id: path.vertices[0]._id,
                        _key: path.vertices[0]._key,
                        router_id: path.vertices[0].router_id,
                        ipv4_address: path.vertices[0].ipv4_address,
                        ipv6_address: path.vertices[0].ipv6_address,
                        prefix: path.vertices[0].prefix,
                        prefix_len: path.vertices[0].prefix_len,
                        name: path.vertices[0].name,
                        sids: path.vertices[0].sids
                    }},
                    destination_info: {{
                        _id: LAST(path.vertices)._id,
                        _key: LAST(path.vertices)._key,
                        router_id: LAST(path.vertices).router_id,
                        ipv4_address: LAST(path.vertices).ipv4_address,
                        ipv6_address: LAST(path.vertices).ipv6_address,
                        prefix: LAST(path.vertices).prefix,
                        prefix_len: LAST(path.vertices).prefix_len,
                        name: LAST(path.vertices).name,
                        sids: LAST(path.vertices).sids
                    }},
                    average_load: FIRST(avg_load)
                }}
            """
        
        # Prepare bind variables
        bind_vars = {
            'source': source,
            'destination': destination,
            'graph_name': collection_name
        }
        
        if algo != 0:
            bind_vars['algo'] = algo
        
        cursor = db.aql.execute(aql, bind_vars=bind_vars)
        
        results = [doc for doc in cursor]
        
        if not results or not results[0]['path']:
            return {
                "found": False,
                "message": "No path found between specified nodes"
            }
        
        # Get the existing response
        response = {
            "found": True,
            "path": results[0]['path'],
            "hopcount": results[0]['hopcount'],
            "vertex_count": results[0]['vertex_count'],
            "source_info": results[0]['source_info'],
            "destination_info": results[0]['destination_info'],
            "direction": direction,
            "average_load": results[0]['average_load'],
            "algo": algo
        }
        
        # Process and append the SRv6 data with algo filtering
        srv6_data = process_path_data(results[0]['path'], source, destination, algo=algo)
        response["srv6_data"] = srv6_data
        
        # Get database connection
        db = get_db()
        
        # Process load data with db connection
        load_data = process_load_data(results[0]['path'], collection_name, db)
        response["load_data"] = load_data
        
        return response
        
    except Exception as e:
        print(f"Error finding shortest path with load weight: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/graphs/{collection_name}/shortest_path/best-paths")
async def get_best_paths(
    collection_name: str,
    source: str,
    destination: str,
    limit: int = 4,
    direction: str = "outbound",
    algo: int = 0  # Flex-Algo to use for SRv6 SID selection
):
    """
    Find multiple best paths between source and destination nodes.
    Default limit is 4 paths, but user can specify more or fewer.
    
    Args:
        algo: Flex-Algo ID for path computation and SRv6 SID selection (default: 0)
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Build algo filtering if needed
        if algo == 0:
            algo_filter = ""
        else:
            algo_filter = f"""
            LET igp_nodes_in_path = (
                FOR v IN p.vertices
                    FILTER CONTAINS(v._id, 'igp_node')
                    RETURN v
            )
            LET nodes_with_algo = (
                FOR node IN igp_nodes_in_path
                    FILTER HAS(node, 'sids') AND node.sids != null
                    FILTER LENGTH(
                        FOR sid IN node.sids
                            FILTER HAS(sid, 'srv6_endpoint_behavior')
                            FILTER HAS(sid.srv6_endpoint_behavior, 'algo')
                            FILTER sid.srv6_endpoint_behavior.algo == {algo}
                            RETURN sid
                    ) > 0
                    RETURN node
            )
            FILTER LENGTH(igp_nodes_in_path) == LENGTH(nodes_with_algo)
            """
        
        # AQL query to get multiple paths
        aql = f"""
        FOR p IN {direction.upper()} k_shortest_paths
            '{source}' TO '{destination}'
            {collection_name}
            OPTIONS {{uniqueVertices: "path", bfs: true}}
            {algo_filter}
            LIMIT {limit}
            RETURN {{
                path: (
                    FOR v IN p.vertices
                    RETURN {{
                        vertex: {{
                            _id: v._id,
                            _key: v._key,
                            name: v.name,
                            sids: v.sids
                        }}
                    }}
                ),
                countries_traversed: p.edges[*].country_codes[*],
                hopcount: LENGTH(p.vertices) - 1,
                vertex_count: LENGTH(p.vertices),
                source_info: FIRST(p.vertices),
                destination_info: LAST(p.vertices)
            }}
        """
        
        cursor = db.aql.execute(aql)
        results = [doc for doc in cursor]
        
        if not results:
            return {
                "found": False,
                "message": "No paths found between specified nodes"
            }
        
        # Process each path and create response
        paths = []
        for result in results:
            path_response = {
                "path": result['path'],
                "hopcount": result['hopcount'],
                "vertex_count": result['vertex_count'],
                "source_info": result['source_info'],
                "destination_info": result['destination_info'],
                "countries_traversed": result['countries_traversed']
            }
            
            # Process and append SRv6 data for each path with algo filtering
            srv6_data = process_path_data(result['path'], source, destination, algo=algo)
            path_response["srv6_data"] = srv6_data
            paths.append(path_response)
        
        return {
            "found": True,
            "total_paths_found": len(paths),
            "direction": direction,
            "algo": algo,
            "paths": paths
        }
        
    except Exception as e:
        print(f"Error finding best paths: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/graphs/{collection_name}/shortest_path/next-best-path")
async def get_next_best_paths(
    collection_name: str,
    source: str,
    destination: str,
    same_hop_limit: int = 4,
    plus_one_limit: int = 8,
    direction: str = "outbound",
    algo: int = 0  # Flex-Algo to use for SRv6 SID selection
):
    """
    Find the shortest path and alternative paths with similar hop counts.
    Allows customization of how many paths to return for each hop count.
    
    Args:
        algo: Flex-Algo ID for path computation and SRv6 SID selection (default: 0)
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Debug prints for shortest path
        print(f"\nProcessing next-best-path request:")
        print(f"Source: {source}")
        print(f"Destination: {destination}")
        print(f"Direction: {direction}")
        
        # First query: Get shortest path and its hop count
        shortest_path_query = f"""
        WITH igp_node
        LET path = (
            FOR v, e IN {direction}
                SHORTEST_PATH '{source}' TO '{destination}'
                {collection_name}
                RETURN {{
                    vertex: {{
                        _id: v._id,
                        _key: v._key,
                        router_id: v.router_id,
                        prefix: v.prefix,
                        name: v.name,
                        sids: v.sids
                    }},
                    edge: e ? {{
                        _id: e._id,
                        _key: e._key,
                        _from: e._from,
                        _to: e._to,
                        latency: e.latency,
                        percent_util_out: e.percent_util_out,
                        country_codes: e.country_codes,
                        load: e.load
                    }} : null
                }}
        )
        RETURN {{
            path: path,
            hopcount: LENGTH(path) - 1
            
        }}
        """
        
        cursor = db.aql.execute(shortest_path_query)
        results = [doc for doc in cursor]
        
        if not results:
            return {
                "found": False,
                "message": "No path found between specified nodes"
            }
            
        shortest_result = results[0]
        base_hopcount = shortest_result['hopcount']
        print(f"Found shortest path with {base_hopcount} hops")
        
        # Second query: Get alternative paths with same hop count
        same_hop_query = f"""
        WITH igp_node
        FOR v, e, p IN {base_hopcount}..{base_hopcount} {direction.upper()}
            '{source}' {collection_name}
            OPTIONS {{ uniquePaths: true, bfs: true }}
            FILTER v._id == '{destination}'
            LIMIT {same_hop_limit}
            RETURN {{
                path: (
                    FOR vertex IN p.vertices
                    RETURN {{
                        vertex: vertex
                    }}
                ),
                hopcount: LENGTH(p.vertices) - 1
            }}
        """
        
        # Third query: Get paths with hop count + 1
        plus_one_hop_query = f"""
        WITH igp_node
        FOR v, e, p IN {base_hopcount + 1}..{base_hopcount + 1} {direction.upper()}
            '{source}' {collection_name}
            OPTIONS {{ uniquePaths: true, bfs: true }}
            FILTER v._id == '{destination}'
            LIMIT {plus_one_limit}
            RETURN {{
                path: (
                    FOR vertex IN p.vertices
                    RETURN {{
                        vertex: vertex
                    }}
                ),
                hopcount: LENGTH(p.vertices) - 1
            }}
        """
        
        # Debug prints
        print(f"\nProcessing next-best-path request:")
        print(f"Source: {source}")
        print(f"Destination: {destination}")
        print(f"Direction: {direction}")
        print(f"Found shortest path with {base_hopcount} hops")
        
        # Execute same hop query
        print(f"\nSearching for paths with same hop count ({base_hopcount})...")
        same_hop_cursor = db.aql.execute(same_hop_query)
        same_hop_paths = [doc for doc in same_hop_cursor]
        print(f"Found {len(same_hop_paths)} alternative paths with {base_hopcount} hops")
        
        # Execute plus one hop query
        print(f"\nSearching for paths with hop count + 1 ({base_hopcount + 1})...")
        plus_one_cursor = db.aql.execute(plus_one_hop_query)
        plus_one_paths = [doc for doc in plus_one_cursor]
        print(f"Found {len(plus_one_paths)} paths with {base_hopcount + 1} hops")
        
        # Process SRv6 data for all paths with algo filtering
        shortest_srv6 = process_path_data(shortest_result['path'], source, destination, algo=algo)
        same_hop_srv6_data = [
            process_path_data(path['path'], source, destination, algo=algo)
            for path in same_hop_paths
        ]
        plus_one_srv6_data = [
            process_path_data(path['path'], source, destination, algo=algo)
            for path in plus_one_paths
        ]
        
        return {
            "found": True,
            "algo": algo,
            "shortest_path": {
                "path": shortest_result['path'],
                "hopcount": shortest_result['hopcount'],
                "srv6_data": shortest_srv6
            },
            "same_hopcount_paths": [{
                "path": path['path'],
                "hopcount": path['hopcount'],
                "srv6_data": srv6
            } for path, srv6 in zip(same_hop_paths, same_hop_srv6_data)],
            "plus_one_hopcount_paths": [{
                "path": path['path'],
                "hopcount": path['hopcount'],
                "srv6_data": srv6
            } for path, srv6 in zip(plus_one_paths, plus_one_srv6_data)],
            "summary": {
                "base_hopcount": base_hopcount,
                "same_hopcount_alternatives": len(same_hop_paths),
                "plus_one_hopcount_alternatives": len(plus_one_paths)
            }
        }
        
    except Exception as e:
        print(f"Error finding next best paths: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/graphs/{collection_name}/traverse")
async def traverse_graph(
    collection_name: str,
    source: str,
    destination: str = None,
    min_depth: int = 1,
    max_depth: int = 4,
    direction: str = "outbound"  # or "inbound", "any"
):
    """
    Traverse graph from a source node with optional destination filtering
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Validate direction parameter
        if direction.lower() not in ["outbound", "inbound", "any"]:
            raise HTTPException(
                status_code=400,
                detail="Direction must be 'outbound', 'inbound', or 'any'"
            )
        
        # Build filter clause if destination node is specified
        filter_clause = f"FILTER v._id == '{destination}'" if destination else ""
        
        # AQL query for traversal with detailed information
        aql = f"""
        LET paths = (
            FOR v, e, p IN {min_depth}..{max_depth} {direction.upper()} 
                '{source}' 
                {collection_name}
                OPTIONS {{uniqueVertices: "path", bfs: true}}
                {filter_clause}
                RETURN DISTINCT {{
                    path: p.vertices[*]._key,
                    sids: p.vertices[*].sids[0].srv6_sid,
                    country_codes: p.edges[*].country_codes,
                    metrics: {{
                        total_latency: SUM(p.edges[*].unidir_link_delay),
                        avg_util: AVG(p.edges[*].percent_util_out),
                        load: AVG(p.edges[*].load),
                        hop_count: LENGTH(p.vertices) - 1
                    }},
                    vertices: (
                        FOR vertex IN p.vertices
                        RETURN {{
                            _id: vertex._id,
                            _key: vertex._key,
                            router_id: vertex.router_id,
                            prefix: vertex.prefix,
                            name: vertex.name,
                            sids: vertex.sids[0].srv6_sid
                        }}
                    ),
                    edges: (
                        FOR edge IN p.edges
                        RETURN {{
                            _key: edge._key,
                            latency: edge.unidir_link_delay,
                            percent_util: edge.percent_util_out,
                            load: edge.load,
                            country_codes: edge.country_codes
                        }}
                    )
                }}
        )
        RETURN {{
            paths: paths,
            total_paths: LENGTH(paths)
        }}
        """
        
        cursor = db.aql.execute(aql)
        result = [doc for doc in cursor][0]  # Get the first (and only) result
        
        return {
            "source": source,
            "destination": destination,
            "min_depth": min_depth,
            "max_depth": max_depth,
            "direction": direction,
            "traversal_results": result['paths'],
            "total_paths": result['total_paths']
        }
        
    except Exception as e:
        print(f"Error traversing graph: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/graphs/{collection_name}/traverse/simple")
async def traverse_graph_simple(
    collection_name: str,
    source: str,
    destination: str = None,
    min_depth: int = 1,
    max_depth: int = 5,
    direction: str = "any"  # or "inbound", "outbound"
):
    """
    Simplified graph traversal with basic path information
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Validate direction parameter
        if direction.lower() not in ["outbound", "inbound", "any"]:
            raise HTTPException(
                status_code=400,
                detail="Direction must be 'outbound', 'inbound', or 'any'"
            )
        
        # Build filter clause if destination node is specified
        filter_clause = f"FILTER v._id == '{destination}'" if destination else ""
        
        # AQL query for simplified traversal
        aql = f"""
        LET paths = (
            FOR v, e, p IN {min_depth}..{max_depth} {direction.upper()} 
                '{source}' 
                {collection_name}
                OPTIONS {{uniqueVertices: "path", bfs: true}}
                {filter_clause}
                RETURN DISTINCT {{
                    path: p.vertices[*]._key,
                    sids: p.vertices[*].sids[0].srv6_sid,
                    country_codes: p.edges[*].country_codes,
                    metrics: {{
                        total_latency: SUM(p.edges[*].unidir_link_delay),
                        avg_util: AVG(p.edges[*].percent_util_out),
                        load: AVG(p.edges[*].load),
                        hop_count: LENGTH(p.vertices) - 1
                    }}
                }}
        )
        RETURN {{
            paths: paths,
            total_paths: LENGTH(paths)
        }}
        """
        
        cursor = db.aql.execute(aql)
        result = [doc for doc in cursor][0]  # Get the first (and only) result
        
        return {
            "source": source,
            "destination": destination,
            "min_depth": min_depth,
            "max_depth": max_depth,
            "direction": direction,
            "traversal_results": result['paths'],
            "total_paths": result['total_paths']
        }
        
    except Exception as e:
        print(f"Error in simple traversal: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/graphs/{collection_name}/neighbors")
async def get_neighbors(
    collection_name: str,
    source: str,
    direction: str = "outbound",  # or "inbound", "any"
    depth: int = 1
):
    """
    Get immediate neighbors of a node
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Validate direction parameter
        if direction.lower() not in ["outbound", "inbound", "any"]:
            raise HTTPException(
                status_code=400,
                detail="Direction must be 'outbound', 'inbound', or 'any'"
            )
        
        # AQL query for neighbors
        aql = f"""
        FOR v, e, p IN 1..{depth} {direction.upper()}
            '{source}'
            {collection_name}
            OPTIONS {{uniqueVertices: "path"}}
            RETURN DISTINCT {{
                neighbor: {{
                    _id: v._id,
                    _key: v._key,
                    router_id: v.router_id,
                    prefix: v.prefix,
                    name: v.name,
                    sids: v.sids[0].srv6_sid
                }},
                edge: {{
                    _key: e._key,
                    latency: e.unidir_link_delay,
                    percent_util: e.percent_util_out,
                    load: e.load,
                    country_codes: e.country_codes
                }},
                metrics: {{
                    hop_count: LENGTH(p.vertices) - 1
                }}
            }}
        """
        
        cursor = db.aql.execute(aql)
        results = [doc for doc in cursor]
        
        return {
            "source": source,
            "direction": direction,
            "depth": depth,
            "neighbor_count": len(results),
            "neighbors": results
        }
        
    except Exception as e:
        print(f"Error getting neighbors: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

# Add this at the bottom of the file
print("\nRegistered routes in graphs.py:")
for route in router.routes:
    print(f"  {route.methods} {route.path}")

# Test route to verify routing is working
@router.get("/api/v1/test")
async def test_route():
    return {"message": "Test route working"} 