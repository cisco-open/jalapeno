from fastapi import APIRouter, HTTPException
from typing import List, Optional, Dict, Any
from arango import ArangoClient
from ..config.settings import Settings
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

router = APIRouter()
settings = Settings()

# Debug print to see registered routes
print("Available VPN routes:")
for route in router.routes:
    print(f"  {route.path}")

# VPN-related collections
VPN_COLLECTIONS = {
    'prefixes': [
        'l3vpn_v4_prefix',
        'l3vpn_v6_prefix'
    ],
    'related': [
        'igp_node',  # PE routers
        'bgp_node'   # PE routers in BGP context
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
# VPN Routes
###################

@router.get("/vpns")
async def get_vpn_collections():
    """
    Get a list of VPN-related collections in the database
    """
    try:
        db = get_db()
        # Get all collections
        collections = db.collections()
        
        # Filter for VPN collections
        vpn_collections = [
            {
                'name': c['name'],
                'type': c['type'],
                'status': c['status'],
                'count': db.collection(c['name']).count()
            }
            for c in collections
            if not c['name'].startswith('_') and 
               (c['name'] in VPN_COLLECTIONS['prefixes'] or 
                c['name'].startswith('l3vpn_') or 
                c['name'].startswith('vpn_'))
        ]
        
        # Sort by name
        vpn_collections.sort(key=lambda x: x['name'])
        
        return {
            'collections': vpn_collections,
            'total_count': len(vpn_collections)
        }
    except Exception as e:
        logger.error(f"Error in get_vpn_collections: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/vpns/{collection_name}")
async def get_vpn_collection_info(collection_name: str):
    """
    Get information about a specific VPN collection
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Verify it's a VPN collection
        if not (collection_name in VPN_COLLECTIONS['prefixes'] or 
                collection_name.startswith('l3vpn_') or 
                collection_name.startswith('vpn_')):
            raise HTTPException(
                status_code=400,
                detail=f"Collection {collection_name} is not a VPN collection"
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
        logger.error(f"Error in get_vpn_collection_info: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/vpns/{collection_name}/summary")
async def get_vpn_summary(collection_name: str):
    """
    Get summary statistics for a VPN collection
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Verify it's a VPN collection
        if not (collection_name in VPN_COLLECTIONS['prefixes'] or 
                collection_name.startswith('l3vpn_') or 
                collection_name.startswith('vpn_')):
            raise HTTPException(
                status_code=400,
                detail=f"Collection {collection_name} is not a VPN collection"
            )
        
        # Get summary statistics based on the actual data structure
        aql = f"""
        LET total_count = LENGTH({collection_name})
        
        LET unique_rds = (
            FOR doc IN {collection_name}
                COLLECT rd = doc.vpn_rd
                RETURN rd
        )
        
        LET unique_route_targets = (
            FOR doc IN {collection_name}
                FOR rt IN doc.base_attrs.ext_community_list
                FILTER STARTS_WITH(rt, 'rt=')
                COLLECT target = rt
                RETURN target
        )
        
        LET unique_nexthops = (
            FOR doc IN {collection_name}
                COLLECT nexthop = doc.nexthop
                RETURN nexthop
        )
        
        LET unique_peer_asns = (
            FOR doc IN {collection_name}
                COLLECT asn = doc.peer_asn
                RETURN asn
        )
        
        LET unique_labels = (
            FOR doc IN {collection_name}
                FOR label IN doc.labels
                COLLECT l = label
                RETURN l
        )
        
        RETURN {{
            total_prefixes: total_count,
            unique_rd_count: LENGTH(unique_rds),
            unique_route_target_count: LENGTH(unique_route_targets),
            unique_nexthop_count: LENGTH(unique_nexthops),
            unique_peer_asn_count: LENGTH(unique_peer_asns),
            unique_label_count: LENGTH(unique_labels)
        }}
        """
        
        cursor = db.aql.execute(aql)
        results = [doc for doc in cursor]
        
        if not results:
            return {
                'collection': collection_name,
                'total_prefixes': 0,
                'unique_rd_count': 0,
                'unique_route_target_count': 0,
                'unique_nexthop_count': 0,
                'unique_peer_asn_count': 0,
                'unique_label_count': 0
            }
        
        return {
            'collection': collection_name,
            **results[0]
        }
        
    except Exception as e:
        logger.error(f"Error in get_vpn_summary: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/vpns/{collection_name}/pe-routers")
async def get_pe_routers(collection_name: str):
    """
    Get a list of PE routers (nexthops) and their prefix counts
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Verify it's a VPN prefix collection
        if collection_name not in VPN_COLLECTIONS['prefixes']:
            raise HTTPException(
                status_code=400,
                detail=f"Collection {collection_name} is not a VPN prefix collection"
            )
        
        # Get PE routers (nexthops) and their prefix counts
        aql = f"""
        FOR doc IN {collection_name}
            COLLECT nexthop = doc.nexthop WITH COUNT INTO count
            RETURN {{
                pe_router: nexthop,
                prefix_count: count
            }}
        """
        
        cursor = db.aql.execute(aql)
        results = [doc for doc in cursor]
        
        return {
            'collection': collection_name,
            'total_pe_routers': len(results),
            'pe_routers': results
        }
        
    except Exception as e:
        logger.error(f"Error in get_pe_routers: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/vpns/{collection_name}/route-targets")
async def get_route_targets(collection_name: str):
    """
    Get a list of route targets and their prefix counts
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Verify it's a VPN prefix collection
        if collection_name not in VPN_COLLECTIONS['prefixes']:
            raise HTTPException(
                status_code=400,
                detail=f"Collection {collection_name} is not a VPN prefix collection"
            )
        
        # Get route targets and their prefix counts
        aql = f"""
        FOR doc IN {collection_name}
            FOR rt IN doc.base_attrs.ext_community_list
                FILTER STARTS_WITH(rt, 'rt=')
                LET clean_rt = SUBSTRING(rt, 3)
                COLLECT route_target = clean_rt WITH COUNT INTO count
                RETURN {{
                    route_target: route_target,
                    prefix_count: count
                }}
        """
        
        cursor = db.aql.execute(aql)
        results = [doc for doc in cursor]
        
        return {
            'collection': collection_name,
            'total_route_targets': len(results),
            'route_targets': results
        }
        
    except Exception as e:
        logger.error(f"Error in get_route_targets: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/vpns/{collection_name}/prefixes/by-pe")
async def get_vpn_prefixes_by_pe(
    collection_name: str,
    pe_router: str,
    limit: int = 100
):
    """
    Get VPN prefixes advertised by a specific PE router (nexthop)
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Verify it's a VPN prefix collection
        if collection_name not in VPN_COLLECTIONS['prefixes']:
            raise HTTPException(
                status_code=400,
                detail=f"Collection {collection_name} is not a VPN prefix collection"
            )
        
        # Get prefixes for the specified PE router (nexthop)
        aql = f"""
        FOR doc IN {collection_name}
            FILTER doc.nexthop == @pe_router
            LIMIT {limit}
            RETURN {{
                _key: doc._key,
                prefix: doc.prefix,
                prefix_len: doc.prefix_len,
                vpn_rd: doc.vpn_rd,
                nexthop: doc.nexthop,
                labels: doc.labels,
                peer_asn: doc.peer_asn,
                route_targets: (
                    FOR rt IN doc.base_attrs.ext_community_list
                    FILTER STARTS_WITH(rt, 'rt=')
                    RETURN SUBSTRING(rt, 3)
                ),
                srv6_sid: doc.prefix_sid.srv6_l3_service.sub_tlvs["1"][0].sid
            }}
        """
        
        cursor = db.aql.execute(aql, bind_vars={'pe_router': pe_router})
        results = [doc for doc in cursor]
        
        # Convert labels to hex in Python and rename to 'function'
        for doc in results:
            if 'labels' in doc and doc['labels']:
                # Convert to hex, trim trailing zeros, and ensure it's at least 4 characters (16 bits)
                doc['function'] = [
                    format(label, 'x').rstrip('0') or '0'  # If all zeros were stripped, return '0'
                    for label in doc['labels']
                ]
                
                # Ensure each function value is at least 4 characters (16 bits)
                doc['function'] = [
                    f if len(f) >= 4 else f.zfill(4)
                    for f in doc['function']
                ]
                
                # Create the combined SID field
                if 'srv6_sid' in doc and doc['srv6_sid'] and doc['function']:
                    # Get the base SRv6 SID
                    base_sid = doc['srv6_sid']
                    # Remove trailing colons if present
                    if base_sid.endswith('::'):
                        base_sid = base_sid[:-2]
                    elif base_sid.endswith(':'):
                        base_sid = base_sid[:-1]
                    
                    # Create the combined SID for each function
                    doc['sid'] = [f"{base_sid}:{func}::" for func in doc['function']]
        
        # Get total count
        aql_count = f"""
        FOR doc IN {collection_name}
            FILTER doc.nexthop == @pe_router
            COLLECT AGGREGATE count = COUNT()
            RETURN count
        """
        
        count_cursor = db.aql.execute(aql_count, bind_vars={'pe_router': pe_router})
        total_count = [count for count in count_cursor][0]
        
        return {
            'collection': collection_name,
            'pe_router': pe_router,
            'total_prefixes': total_count,
            'prefixes': results,
            'limit_applied': limit
        }
        
    except Exception as e:
        logger.error(f"Error in get_vpn_prefixes_by_pe: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/vpns/{collection_name}/prefixes/by-rt")
async def get_vpn_prefixes_by_rt(
    collection_name: str,
    route_target: str,
    limit: int = 100
):
    """
    Get VPN prefixes associated with a specific route target
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Verify it's a VPN prefix collection
        if collection_name not in VPN_COLLECTIONS['prefixes']:
            raise HTTPException(
                status_code=400,
                detail=f"Collection {collection_name} is not a VPN prefix collection"
            )
        
        # Format the route target to match how it's stored
        formatted_rt = f"rt={route_target}"
        
        # Get prefixes for the specified route target
        aql = f"""
        FOR doc IN {collection_name}
            FILTER @route_target IN doc.base_attrs.ext_community_list
            LIMIT {limit}
            RETURN {{
                _key: doc._key,
                prefix: doc.prefix,
                prefix_len: doc.prefix_len,
                vpn_rd: doc.vpn_rd,
                nexthop: doc.nexthop,
                labels: doc.labels,
                peer_asn: doc.peer_asn,
                route_targets: (
                    FOR rt IN doc.base_attrs.ext_community_list
                    FILTER STARTS_WITH(rt, 'rt=')
                    RETURN SUBSTRING(rt, 3)
                ),
                srv6_sid: doc.prefix_sid.srv6_l3_service.sub_tlvs["1"][0].sid
            }}
        """
        
        cursor = db.aql.execute(aql, bind_vars={'route_target': formatted_rt})
        results = [doc for doc in cursor]
        
        # Convert labels to hex in Python and rename to 'function'
        for doc in results:
            if 'labels' in doc and doc['labels']:
                # Convert to hex, trim trailing zeros, and ensure it's at least 4 characters (16 bits)
                doc['function'] = [
                    format(label, 'x').rstrip('0') or '0'  # If all zeros were stripped, return '0'
                    for label in doc['labels']
                ]
                
                # Ensure each function value is at least 4 characters (16 bits)
                doc['function'] = [
                    f if len(f) >= 4 else f.zfill(4)
                    for f in doc['function']
                ]
                
                # Create the combined SID field
                if 'srv6_sid' in doc and doc['srv6_sid'] and doc['function']:
                    # Get the base SRv6 SID
                    base_sid = doc['srv6_sid']
                    # Remove trailing colons if present
                    if base_sid.endswith('::'):
                        base_sid = base_sid[:-2]
                    elif base_sid.endswith(':'):
                        base_sid = base_sid[:-1]
                    
                    # Create the combined SID for each function
                    doc['sid'] = [f"{base_sid}:{func}::" for func in doc['function']]
        
        # Get total count
        aql_count = f"""
        FOR doc IN {collection_name}
            FILTER @route_target IN doc.base_attrs.ext_community_list
            COLLECT AGGREGATE count = COUNT()
            RETURN count
        """
        
        count_cursor = db.aql.execute(aql_count, bind_vars={'route_target': formatted_rt})
        total_count = [count for count in count_cursor][0]
        
        # Group by nexthop for summary
        nexthop_summary = {}
        for prefix in results:
            nexthop = prefix['nexthop']
            if nexthop not in nexthop_summary:
                nexthop_summary[nexthop] = 0
            nexthop_summary[nexthop] += 1
        
        nexthop_list = [{"nexthop": nh, "prefix_count": count} for nh, count in nexthop_summary.items()]
        
        return {
            'collection': collection_name,
            'route_target': route_target,
            'total_prefixes': total_count,
            'advertising_pe_count': len(nexthop_summary),
            'advertising_pes': nexthop_list,
            'prefixes': results,
            'limit_applied': limit
        }
        
    except Exception as e:
        logger.error(f"Error in get_vpn_prefixes_by_rt: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/vpns/{collection_name}/prefixes/search")
async def search_vpn_prefixes(
    collection_name: str,
    prefix: Optional[str] = None,
    prefix_exact: Optional[bool] = False,
    route_target: Optional[str] = None,
    vpn_rd: Optional[str] = None,
    limit: int = 100
):
    """
    Search for VPN prefixes with flexible filtering options.
    
    Parameters:
    - prefix: Search for this prefix (can be partial match if prefix_exact=False)
    - prefix_exact: If True, match the prefix exactly; if False, use prefix as a substring
    - route_target: Filter by this route target
    - vpn_rd: Filter by this VPN Route Distinguisher
    - limit: Maximum number of results to return
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Verify it's a VPN prefix collection
        if collection_name not in VPN_COLLECTIONS['prefixes']:
            raise HTTPException(
                status_code=400,
                detail=f"Collection {collection_name} is not a VPN prefix collection"
            )
        
        # Build filter conditions based on provided parameters
        filter_conditions = []
        bind_vars = {}
        
        if prefix:
            bind_vars['prefix'] = prefix
            if prefix_exact:
                filter_conditions.append("doc.prefix == @prefix")
            else:
                filter_conditions.append("CONTAINS(doc.prefix, @prefix)")
        
        if route_target:
            formatted_rt = f"rt={route_target}"
            bind_vars['route_target'] = formatted_rt
            filter_conditions.append("@route_target IN doc.base_attrs.ext_community_list")
        
        if vpn_rd:
            bind_vars['vpn_rd'] = vpn_rd
            filter_conditions.append("doc.vpn_rd == @vpn_rd")
        
        # If no filters provided, return an error
        if not filter_conditions:
            raise HTTPException(
                status_code=400,
                detail="At least one search parameter (prefix, route_target, or vpn_rd) must be provided"
            )
        
        # Combine filter conditions
        filter_clause = " AND ".join(filter_conditions)
        
        # Get matching prefixes
        aql = f"""
        FOR doc IN {collection_name}
            FILTER {filter_clause}
            LIMIT {limit}
            RETURN {{
                _key: doc._key,
                prefix: doc.prefix,
                prefix_len: doc.prefix_len,
                vpn_rd: doc.vpn_rd,
                nexthop: doc.nexthop,
                labels: doc.labels,
                peer_asn: doc.peer_asn,
                route_targets: (
                    FOR rt IN doc.base_attrs.ext_community_list
                    FILTER STARTS_WITH(rt, 'rt=')
                    RETURN SUBSTRING(rt, 3)
                ),
                srv6_sid: doc.prefix_sid.srv6_l3_service.sub_tlvs["1"][0].sid
            }}
        """
        
        cursor = db.aql.execute(aql, bind_vars=bind_vars)
        results = [doc for doc in cursor]
        
        # Convert labels to hex in Python and rename to 'function'
        for doc in results:
            if 'labels' in doc and doc['labels']:
                # Convert to hex, trim trailing zeros, and ensure it's at least 4 characters (16 bits)
                doc['function'] = [
                    format(label, 'x').rstrip('0') or '0'  # If all zeros were stripped, return '0'
                    for label in doc['labels']
                ]
                
                # Ensure each function value is at least 4 characters (16 bits)
                doc['function'] = [
                    f if len(f) >= 4 else f.zfill(4)
                    for f in doc['function']
                ]
                
                # Create the combined SID field
                if 'srv6_sid' in doc and doc['srv6_sid'] and doc['function']:
                    # Get the base SRv6 SID
                    base_sid = doc['srv6_sid']
                    # Remove trailing colons if present
                    if base_sid.endswith('::'):
                        base_sid = base_sid[:-2]
                    elif base_sid.endswith(':'):
                        base_sid = base_sid[:-1]
                    
                    # Create the combined SID for each function
                    doc['sid'] = [f"{base_sid}:{func}::" for func in doc['function']]
        
        # Get total count
        aql_count = f"""
        FOR doc IN {collection_name}
            FILTER {filter_clause}
            COLLECT AGGREGATE count = COUNT()
            RETURN count
        """
        
        count_cursor = db.aql.execute(aql_count, bind_vars=bind_vars)
        total_count = [count for count in count_cursor][0]
        
        # Build response with search criteria
        search_criteria = {}
        if prefix:
            search_criteria['prefix'] = prefix
            search_criteria['prefix_exact'] = prefix_exact
        if route_target:
            search_criteria['route_target'] = route_target
        if vpn_rd:
            search_criteria['vpn_rd'] = vpn_rd
        
        return {
            'collection': collection_name,
            'search_criteria': search_criteria,
            'total_prefixes': total_count,
            'prefixes': results,
            'limit_applied': limit
        }
        
    except Exception as e:
        logger.error(f"Error in search_vpn_prefixes: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

@router.get("/vpns/{collection_name}/prefixes/by-pe-rt")
async def get_vpn_prefixes_by_pe_rt(
    collection_name: str,
    pe_router: str,
    route_target: str,
    limit: int = 100
):
    """
    Get VPN prefixes that match both a specific PE router (nexthop) and route target.
    
    Parameters:
    - pe_router: The PE router's nexthop address
    - route_target: The route target to filter by
    - limit: Maximum number of results to return
    """
    try:
        db = get_db()
        if not db.has_collection(collection_name):
            raise HTTPException(
                status_code=404,
                detail=f"Collection {collection_name} not found"
            )
        
        # Verify it's a VPN prefix collection
        if collection_name not in VPN_COLLECTIONS['prefixes']:
            raise HTTPException(
                status_code=400,
                detail=f"Collection {collection_name} is not a VPN prefix collection"
            )
        
        # Format the route target to match how it's stored
        formatted_rt = f"rt={route_target}"
        
        # Get prefixes matching both PE router and route target
        aql = f"""
        FOR doc IN {collection_name}
            FILTER doc.nexthop == @pe_router
            FILTER @route_target IN doc.base_attrs.ext_community_list
            LIMIT {limit}
            RETURN {{
                _key: doc._key,
                prefix: doc.prefix,
                prefix_len: doc.prefix_len,
                vpn_rd: doc.vpn_rd,
                nexthop: doc.nexthop,
                labels: doc.labels,
                peer_asn: doc.peer_asn,
                route_targets: (
                    FOR rt IN doc.base_attrs.ext_community_list
                    FILTER STARTS_WITH(rt, 'rt=')
                    RETURN SUBSTRING(rt, 3)
                ),
                srv6_sid: doc.prefix_sid.srv6_l3_service.sub_tlvs["1"][0].sid
            }}
        """
        
        cursor = db.aql.execute(aql, bind_vars={
            'pe_router': pe_router,
            'route_target': formatted_rt
        })
        results = [doc for doc in cursor]
        
        # Convert labels to hex in Python and rename to 'function'
        for doc in results:
            if 'labels' in doc and doc['labels']:
                # Convert to hex, trim trailing zeros, and ensure it's at least 4 characters (16 bits)
                doc['function'] = [
                    format(label, 'x').rstrip('0') or '0'  # If all zeros were stripped, return '0'
                    for label in doc['labels']
                ]
                
                # Ensure each function value is at least 4 characters (16 bits)
                doc['function'] = [
                    f if len(f) >= 4 else f.zfill(4)
                    for f in doc['function']
                ]
                
                # Create the combined SID field
                if 'srv6_sid' in doc and doc['srv6_sid'] and doc['function']:
                    # Get the base SRv6 SID
                    base_sid = doc['srv6_sid']
                    # Remove trailing colons if present
                    if base_sid.endswith('::'):
                        base_sid = base_sid[:-2]
                    elif base_sid.endswith(':'):
                        base_sid = base_sid[:-1]
                    
                    # Create the combined SID for each function
                    doc['sid'] = [f"{base_sid}:{func}::" for func in doc['function']]
        
        # Get total count
        aql_count = f"""
        FOR doc IN {collection_name}
            FILTER doc.nexthop == @pe_router
            FILTER @route_target IN doc.base_attrs.ext_community_list
            COLLECT AGGREGATE count = COUNT()
            RETURN count
        """
        
        count_cursor = db.aql.execute(aql_count, bind_vars={
            'pe_router': pe_router,
            'route_target': formatted_rt
        })
        total_count = [count for count in count_cursor][0]
        
        return {
            'collection': collection_name,
            'pe_router': pe_router,
            'route_target': route_target,
            'total_prefixes': total_count,
            'prefixes': results,
            'limit_applied': limit
        }
        
    except Exception as e:
        logger.error(f"Error in get_vpn_prefixes_by_pe_rt: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=str(e)
        )

# Add this at the bottom of the file
print("\nRegistered routes in vpns.py:")
for route in router.routes:
    print(f"  {route.methods} {route.path}")