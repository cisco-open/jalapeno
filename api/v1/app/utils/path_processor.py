from math import ceil
from typing import List, Dict, Any
import json

def process_path_data(
    path_data: List[Dict[Any, Any]], 
    source: str, 
    destination: str,
    usid_block: str = None,
    algo: int = 0
) -> Dict:
    """
    Process shortest path data to extract SRv6 information
    
    Args:
        path_data: List of path nodes with vertex/edge information
        source: Source node identifier
        destination: Destination node identifier
        usid_block: Optional USID block prefix (e.g., 'fc00:0:', 'fc00:2:', 'fbbb:0:')
                   If None, will auto-detect from the first SID matching the algo
        algo: Flex-Algo ID to filter SIDs (default: 0)
    """
    try:
        
        # Calculate path metrics
        hopcount = len(path_data)
        print(f"Hopcount: {hopcount}, Algo: {algo}")
        
        # Extract SID locators filtered by algo
        locators = []
        for node in path_data:
            # print(f"Processing node: {json.dumps(node, indent=2)}")
            # Check for vertex and sids in the vertex object
            if 'vertex' in node and 'sids' in node['vertex']:
                vertex_sids = node['vertex']['sids']
                if isinstance(vertex_sids, list) and len(vertex_sids) > 0:
                    # Filter SIDs by algo
                    matching_sid = None
                    for sid_entry in vertex_sids:
                        if isinstance(sid_entry, dict):
                            # Check if this SID matches the requested algo
                            if ('srv6_endpoint_behavior' in sid_entry and 
                                'algo' in sid_entry['srv6_endpoint_behavior'] and
                                sid_entry['srv6_endpoint_behavior']['algo'] == algo):
                                matching_sid = sid_entry.get('srv6_sid')
                                break
                    
                    # If we found a matching SID, add it to locators
                    if matching_sid:
                        locators.append(matching_sid)
                        # print(f"Added SID for algo {algo}: {matching_sid}")
        
        print(f"Collected locators for algo {algo}: {locators}")
        
        # Auto-detect USID block from first locator if not provided
        if usid_block is None and len(locators) > 0:
            # Extract the block from the first SID (everything up to and including the second colon)
            first_sid = locators[0]
            parts = first_sid.split(':')
            if len(parts) >= 3:
                # Reconstruct block as first two parts + trailing colon
                usid_block = f"{parts[0]}:{parts[1]}:"
                print(f"Auto-detected USID block: {usid_block}")
            else:
                # Fallback to default if format is unexpected
                usid_block = 'fc00:0:'
                print(f"Could not auto-detect USID block, using default: {usid_block}")
        elif usid_block is None:
            # No locators and no explicit block provided
            usid_block = 'fc00:0:'
            print(f"No locators found, using default USID block: {usid_block}")
        
        # Process USID information
        usid = []
        for sid in locators:
            if sid and usid_block in sid:
                usid_list = sid.split(usid_block)
                sid_value = usid_list[1]
                usid_int = sid_value.split(':')
                usid.append(usid_int[0])
                # print(f"Processed USID: {usid_int[0]}")
        
        # Build SRv6 USID carrier
        sidlist = ":".join(usid) + "::"
        srv6_sid = usid_block + sidlist
        print(f"Final SRv6 SID for algo {algo}: {srv6_sid}")
        
        result = {
            'srv6_sid_list': locators,
            'srv6_usid': srv6_sid,
            'usid_block': usid_block,
            'algo': algo
        }
        # print(f"Returning result: {json.dumps(result, indent=2)}")
        return result
        
    except Exception as e:
        print(f"Error in path_processor: {str(e)}")
        return {
            'error': str(e),
            'srv6_sid_list': [],
            'srv6_usid': ''
        } 