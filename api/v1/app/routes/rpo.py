from fastapi import APIRouter, HTTPException, Query
from typing import List, Optional, Dict, Any, Union
from arango import ArangoClient
from ..config.settings import Settings
import logging
from .graphs import get_shortest_path

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

router = APIRouter()
settings = Settings()

# Supported metric types and their optimization strategies
SUPPORTED_METRICS = {
    'cpu_utilization': {'type': 'numeric', 'optimize': 'minimize'},
    'gpu_utilization': {'type': 'numeric', 'optimize': 'minimize'},
    'memory_utilization': {'type': 'numeric', 'optimize': 'minimize'},
    'time_to_first_token': {'type': 'numeric', 'optimize': 'minimize'},
    'cost_per_million_tokens': {'type': 'numeric', 'optimize': 'minimize'},
    'cost_per_hour': {'type': 'numeric', 'optimize': 'minimize'},
    'gpu_model': {'type': 'string', 'optimize': 'exact_match'},
    'language_model': {'type': 'string', 'optimize': 'exact_match'},
    'response_time': {'type': 'numeric', 'optimize': 'minimize'}
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

@router.get("/rpo")
async def get_rpo_info():
    """
    Get information about Resource Path Optimization (RPO) capabilities
    """
    try:
        db = get_db()
        
        # Get all collections to identify potential graph collections
        all_collections = db.collections()
        graph_collections = []
        
        for collection in all_collections:
            collection_name = collection['name']
            # Look for collections that might be graph collections
            # Common patterns: *_graph, topology_*, network_*, etc.
            # Exclude vertex collections like igp_domain, igp_node
            if (any(pattern in collection_name.lower() for pattern in ['graph', 'topology', 'network']) and 
                not any(vertex_pattern in collection_name.lower() for vertex_pattern in ['domain', 'node', 'vertex'])):
                graph_collections.append(collection_name)
        
        return {
            'supported_metrics': SUPPORTED_METRICS,
            'description': 'Resource Path Optimization (RPO) API for intelligent destination selection',
            'available_graph_collections': sorted(graph_collections),
            'note': 'Use graphs parameter to specify which topology graph to use for path finding'
        }
        
    except Exception as e:
        logger.warning(f"Could not fetch graph collections: {str(e)}")
        return {
            'supported_metrics': SUPPORTED_METRICS,
            'description': 'Resource Path Optimization (RPO) API for intelligent destination selection',
            'available_graph_collections': [],
            'note': 'Use graphs parameter to specify which topology graph to use for path finding'
        }

@router.get("/rpo/{collection_name}")
async def get_collection_endpoints(
    collection_name: str,
    limit: Optional[int] = None
):
    """
    Get all endpoints from a specific collection with their metrics
    """
    try:
        db = get_db()
        
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Query all endpoints from the collection
        endpoints_query = f"""
        FOR doc IN {collection_name}
            RETURN doc
        """
        
        if limit:
            endpoints_query = f"""
            FOR doc IN {collection_name}
                LIMIT {limit}
                RETURN doc
            """
        
        cursor = db.aql.execute(endpoints_query)
        endpoints = [doc for doc in cursor]
        
        return {
            'collection': collection_name,
            'type': 'collection',
            'count': len(endpoints),
            'data': endpoints
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error getting collection endpoints: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/rpo/{collection_name}/select-optimal")
async def select_optimal_endpoint(
    collection_name: str,
    source: str = Query(..., description="Source endpoint ID"),
    metric: str = Query(..., description="Metric to optimize for"),
    value: Optional[str] = Query(None, description="Required value for exact match metrics"),
    graphs: str = Query(..., description="Graph collection to use for path finding"),
    direction: str = Query("outbound", description="Direction for path finding"),
    algo: Optional[int] = Query(None, description="Flex-Algo ID to use for path finding (default: 0)")
):
    """
    Select optimal destination endpoint from a collection based on metrics for Resource Path Optimization
    """
    try:
        db = get_db()
        
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        if metric not in SUPPORTED_METRICS:
            raise HTTPException(
                status_code=400,
                detail=f"Unsupported metric: {metric}. Supported metrics: {list(SUPPORTED_METRICS.keys())}"
            )
        
        # Get all endpoints from the collection
        endpoints_query = f"""
        FOR doc IN {collection_name}
            RETURN doc
        """
        
        cursor = db.aql.execute(endpoints_query)
        endpoints = [doc for doc in cursor]
        
        if not endpoints:
            raise HTTPException(
                status_code=404,
                detail=f"No endpoints found in collection {collection_name}"
            )
        
        # Filter endpoints with valid metric values
        metric_config = SUPPORTED_METRICS[metric]
        optimization_strategy = metric_config['optimize']
        
        if optimization_strategy == 'exact_match':
            if not value:
                raise HTTPException(
                    status_code=400,
                    detail=f"Value required for exact match metric: {metric}"
                )
            
            # Find endpoints that match the exact value
            valid_endpoints = [
                ep for ep in endpoints
                if ep.get(metric) == value
            ]
            
            if not valid_endpoints:
                raise HTTPException(
                    status_code=404,
                    detail=f"No endpoints found with {metric} = {value}"
                )
            
            selected_endpoint = valid_endpoints[0]
            
        elif optimization_strategy == 'minimize':
            # Find endpoint with minimum value for the metric (excluding null values)
            valid_endpoints = [
                ep for ep in endpoints 
                if ep.get(metric) is not None
            ]
            
            if not valid_endpoints:
                raise HTTPException(
                    status_code=404,
                    detail=f"No endpoints found with valid {metric} values"
                )
            
            selected_endpoint = min(
                valid_endpoints,
                key=lambda x: x.get(metric)
            )
            
        elif optimization_strategy == 'maximize':
            # Find endpoint with maximum value for the metric (excluding null values)
            valid_endpoints = [
                ep for ep in endpoints 
                if ep.get(metric) is not None
            ]
            
            if not valid_endpoints:
                raise HTTPException(
                    status_code=404,
                    detail=f"No endpoints found with valid {metric} values"
                )
            
            selected_endpoint = max(
                valid_endpoints,
                key=lambda x: x.get(metric)
            )
        
        else:
            raise HTTPException(
                status_code=500,
                detail=f"Unknown optimization strategy: {optimization_strategy}"
            )
        
        # Find shortest path to selected endpoint
        destination = selected_endpoint['_id']
        logger.info(f"Finding shortest path from {source} to {destination}...")
        
        try:
            path_result = await get_shortest_path(
                collection_name=graphs,
                source=source,
                destination=destination,
                direction=direction,
                algo=algo
            )
        except Exception as path_error:
            logger.warning(f"Could not find path: {str(path_error)}")
            path_result = {
                "found": False,
                "error": str(path_error),
                "message": "No path found between specified nodes"
            }
        
        return {
            'collection': collection_name,
            'source': source,
            'selected_endpoint': selected_endpoint,
            'optimization_metric': metric,
            'metric_value': selected_endpoint.get(metric),
            'optimization_strategy': optimization_strategy,
            'algo': algo if algo is not None else 0,
            'total_endpoints_evaluated': len(endpoints),
            'valid_endpoints_count': len(valid_endpoints) if 'valid_endpoints' in locals() else len(endpoints),
            'path_result': path_result,
            'summary': {
                'destination': destination,
                'destination_name': selected_endpoint.get('name', 'Unknown'),
                'path_found': path_result.get('found', False),
                'hop_count': path_result.get('hopcount', 0)
            }
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error in select_optimal_endpoint: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/rpo/{collection_name}/select-from-list")
async def select_from_specific_endpoints(
    collection_name: str,
    source: str = Query(..., description="Source endpoint ID"),
    destinations: str = Query(..., description="Comma-separated list of destination endpoint IDs"),
    metric: str = Query(..., description="Metric to optimize for"),
    value: Optional[str] = Query(None, description="Required value for exact match metrics"),
    graphs: str = Query(..., description="Graph collection to use for path finding"),
    direction: str = Query("outbound", description="Direction for path finding"),
    algo: Optional[int] = Query(None, description="Flex-Algo ID to use for path finding (default: 0)")
):
    """
    Select optimal destination from a specific list of endpoints for Resource Path Optimization
    """
    try:
        db = get_db()
        
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        if metric not in SUPPORTED_METRICS:
            raise HTTPException(
                status_code=400,
                detail=f"Unsupported metric: {metric}. Supported metrics: {list(SUPPORTED_METRICS.keys())}"
            )
        
        # Parse destination list
        destination_list = [dest.strip() for dest in destinations.split(',')]
        
        # Get endpoint details for each destination
        endpoints = []
        for dest_id in destination_list:
            # Extract collection and key from dest_id (e.g., "hosts/amsterdam" -> collection="hosts", key="amsterdam")
            if '/' in dest_id:
                dest_collection, key = dest_id.split('/', 1)
            else:
                dest_collection = collection_name
                key = dest_id
            
            # Try to get the endpoint from the specific collection
            if db.has_collection(dest_collection):
                try:
                    endpoint = db.collection(dest_collection).get(key)
                    if endpoint:
                        endpoints.append(endpoint)
                    else:
                        logger.warning(f"Could not find endpoint: {dest_id}")
                except Exception as e:
                    logger.warning(f"Error getting endpoint {dest_id}: {str(e)}")
            else:
                logger.warning(f"Collection {dest_collection} not found for endpoint: {dest_id}")
        
        if not endpoints:
            raise HTTPException(
                status_code=404,
                detail="No valid endpoints found in the provided list"
            )
        
        # Apply selection logic
        metric_config = SUPPORTED_METRICS[metric]
        optimization_strategy = metric_config['optimize']
        
        if optimization_strategy == 'exact_match':
            if not value:
                raise HTTPException(
                    status_code=400,
                    detail=f"Value required for exact match metric: {metric}"
                )
            
            valid_endpoints = [
                ep for ep in endpoints
                if ep.get(metric) == value
            ]
            
            if not valid_endpoints:
                raise HTTPException(
                    status_code=404,
                    detail=f"No endpoints found with {metric} = {value}"
                )
            
            selected_endpoint = valid_endpoints[0]
            
        elif optimization_strategy == 'minimize':
            valid_endpoints = [
                ep for ep in endpoints 
                if ep.get(metric) is not None
            ]
            
            if not valid_endpoints:
                raise HTTPException(
                    status_code=404,
                    detail=f"No endpoints found with valid {metric} values"
                )
            
            selected_endpoint = min(
                valid_endpoints,
                key=lambda x: x.get(metric)
            )
            
        elif optimization_strategy == 'maximize':
            valid_endpoints = [
                ep for ep in endpoints 
                if ep.get(metric) is not None
            ]
            
            if not valid_endpoints:
                raise HTTPException(
                    status_code=404,
                    detail=f"No endpoints found with valid {metric} values"
                )
            
            selected_endpoint = max(
                valid_endpoints,
                key=lambda x: x.get(metric)
            )
        
        # Find shortest path to selected endpoint
        destination = selected_endpoint['_id']
        logger.info(f"Finding shortest path from {source} to {destination}...")
        
        try:
            path_result = await get_shortest_path(
                collection_name=graphs,
                source=source,
                destination=destination,
                direction=direction,
                algo=algo
            )
        except Exception as path_error:
            logger.warning(f"Could not find path: {str(path_error)}")
            path_result = {
                "found": False,
                "error": str(path_error),
                "message": "No path found between specified nodes"
            }
        
        return {
            'collection': collection_name,
            'source': source,
            'selected_endpoint': selected_endpoint,
            'optimization_metric': metric,
            'metric_value': selected_endpoint.get(metric),
            'optimization_strategy': optimization_strategy,
            'algo': algo if algo is not None else 0,
            'total_candidates': len(endpoints),
            'valid_endpoints_count': len(valid_endpoints) if 'valid_endpoints' in locals() else len(endpoints),
            'path_result': path_result,
            'summary': {
                'destination': destination,
                'destination_name': selected_endpoint.get('name', 'Unknown'),
                'path_found': path_result.get('found', False),
                'hop_count': path_result.get('hopcount', 0)
            }
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error in select_from_specific_endpoints: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )