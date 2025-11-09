from typing import List, Dict, Any

def process_load_data(
    path_data: List[Dict[Any, Any]],
    collection_name: str,
    db,
    load_increment: int = 10
) -> Dict:
    """
    Process path data to update and calculate load metrics
    
    Args:
        path_data: List of dictionaries containing path information
        collection_name: Name of the graph collection
        db: Database connection
        load_increment: Amount to increment load by (default: 10)
        
    Returns:
        Dictionary containing load processing results
    """
    try:
        # Update edge documents with load value
        updated_edges = []
        highest_load = 0
        highest_load_edge = None
        
        for doc in path_data:
            if doc.get('edge') and doc['edge'].get('_key'):
                edge_key = doc['edge']['_key']
                # Get current edge document
                edge_doc = db.collection(collection_name).get({'_key': edge_key})
                if edge_doc:
                    # Get current load value, default to 0 if it doesn't exist
                    current_load = edge_doc.get('load', 0)
                    new_load = current_load + load_increment
                    
                    # Track highest load
                    if new_load > highest_load:
                        highest_load = new_load
                        highest_load_edge = edge_key
                        
                    # Update with incremented load
                    db.collection(collection_name).update_match(
                        {'_key': edge_key},
                        {'load': new_load}
                    )
                    updated_edges.append(edge_key)
                    print(f"Load updated for edge: {edge_key}")

        # Calculate average load after updates
        total_load = 0
        edge_count = 0
        updated_loads = []
        
        for doc in path_data:
            if doc.get('edge') and doc['edge'].get('_key'):
                edge_key = doc['edge']['_key']
                edge_doc = db.collection(collection_name).get({'_key': edge_key})
                if edge_doc:
                    current_load = edge_doc.get('load', 0)
                    total_load += current_load
                    edge_count += 1
                    updated_loads.append({
                        'edge_key': edge_key,
                        'load': current_load
                    })

        avg_load = total_load / edge_count if edge_count > 0 else 0
        print(f"Average load across path: {avg_load}")

        return {
            'updated_edges': updated_edges,
            'edge_loads': updated_loads,
            'average_load': avg_load,
            'total_load': total_load,
            'edge_count': edge_count,
            'highest_load': {
                'edge_key': highest_load_edge,
                'load_value': highest_load
            }
        }

    except Exception as e:
        print(f"Error processing load data: {str(e)}")
        return {
            'error': str(e),
            'updated_edges': [],
            'edge_loads': [],
            'average_load': 0,
            'total_load': 0,
            'edge_count': 0,
            'highest_load': {
                'edge_key': None,
                'load_value': 0
            }
        } 