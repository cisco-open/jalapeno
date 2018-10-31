package handler
import (
	"strconv"
)
func check_asn_location(asn string) bool {
	var is_internal_asn bool = false
	current_asn, _ := strconv.Atoi(asn)
	if ((current_asn >= 64512) && (current_asn <= 65534)) || ((current_asn >= 4200000000) && (current_asn <= 4294967294)) {
       		is_internal_asn = true
   	}
	return is_internal_asn
}

// Calculates sid value using initial SRGB label and sid-index
func calculate_sid(sr_beginning_label int, sid_index string) string {
        sid_index_val, _ := strconv.ParseInt(sid_index, 10, 0)
        sid_val := int(sr_beginning_label) + int(sid_index_val)
        sid := strconv.Itoa(int(sid_val))
        return sid
}


// Create a Router document and upsert it into the Routers Collection
//func create_router_document(router_ip, bgp_id, name="", node_sid_index="", sr_gb="", sr_node_sid="", asn="") {
//
//
//}
