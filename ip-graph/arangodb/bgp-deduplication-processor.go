// Copyright (c) 2022-2025 Cisco Systems, Inc. and its affiliates
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//
// The contents of this file are licensed under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with the
// License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package arangodb

import (
	"context"
	"fmt"

	"github.com/golang/glog"
)

// BGPDeduplicationProcessor handles advanced BGP prefix deduplication and classification
type BGPDeduplicationProcessor struct {
	db *arangoDB
}

// NewBGPDeduplicationProcessor creates a new BGP deduplication processor
func NewBGPDeduplicationProcessor(db *arangoDB) *BGPDeduplicationProcessor {
	return &BGPDeduplicationProcessor{
		db: db,
	}
}

// ProcessInitialBGPDeduplication performs initial BGP prefix deduplication and classification
func (bdp *BGPDeduplicationProcessor) ProcessInitialBGPDeduplication(ctx context.Context) error {
	glog.Info("Starting BGP prefix deduplication and classification...")

	// Process IPv4 prefixes
	if err := bdp.processIPv4PrefixDeduplication(ctx); err != nil {
		return fmt.Errorf("failed to process IPv4 prefix deduplication: %w", err)
	}

	// Process IPv6 prefixes
	if err := bdp.processIPv6PrefixDeduplication(ctx); err != nil {
		return fmt.Errorf("failed to process IPv6 prefix deduplication: %w", err)
	}

	glog.Info("BGP prefix deduplication and classification completed successfully")
	return nil
}

// processIPv4PrefixDeduplication handles IPv4 prefix deduplication and classification
func (bdp *BGPDeduplicationProcessor) processIPv4PrefixDeduplication(ctx context.Context) error {
	glog.V(6).Info("Processing IPv4 prefix deduplication...")

	// Clear existing classified prefixes (for reprocessing)
	if err := bdp.clearExistingClassifiedPrefixes(ctx, true); err != nil {
		glog.Warningf("Failed to clear existing IPv4 classified prefixes: %v", err)
	}

	// 1. Process eBGP Private ASN prefixes
	if err := bdp.processEBGPPrivateIPv4Prefixes(ctx); err != nil {
		return fmt.Errorf("failed to process eBGP private IPv4 prefixes: %w", err)
	}

	// 2. Process Internet (public ASN) prefixes
	if err := bdp.processInternetIPv4Prefixes(ctx); err != nil {
		return fmt.Errorf("failed to process Internet IPv4 prefixes: %w", err)
	}

	// 3. Process iBGP prefixes (last, to avoid conflicts with external prefixes)
	if err := bdp.processIBGPIPv4Prefixes(ctx); err != nil {
		return fmt.Errorf("failed to process iBGP IPv4 prefixes: %w", err)
	}

	glog.V(6).Info("IPv4 prefix deduplication completed")
	return nil
}

// processIPv6PrefixDeduplication handles IPv6 prefix deduplication and classification
func (bdp *BGPDeduplicationProcessor) processIPv6PrefixDeduplication(ctx context.Context) error {
	glog.V(6).Info("Processing IPv6 prefix deduplication...")

	// Clear existing classified prefixes (for reprocessing)
	if err := bdp.clearExistingClassifiedPrefixes(ctx, false); err != nil {
		glog.Warningf("Failed to clear existing IPv6 classified prefixes: %v", err)
	}

	// 1. Process eBGP Private ASN prefixes (2-byte)
	if err := bdp.processEBGPPrivateIPv6Prefixes(ctx); err != nil {
		return fmt.Errorf("failed to process eBGP private IPv6 prefixes: %w", err)
	}

	// 2. Process eBGP Private ASN prefixes (4-byte)
	if err := bdp.processEBGP4BytePrivateIPv6Prefixes(ctx); err != nil {
		return fmt.Errorf("failed to process eBGP 4-byte private IPv6 prefixes: %w", err)
	}

	// 3. Process Internet (public ASN) prefixes
	if err := bdp.processInternetIPv6Prefixes(ctx); err != nil {
		return fmt.Errorf("failed to process Internet IPv6 prefixes: %w", err)
	}

	// 4. Process iBGP prefixes (last, to avoid conflicts with external prefixes)
	if err := bdp.processIBGPIPv6Prefixes(ctx); err != nil {
		return fmt.Errorf("failed to process iBGP IPv6 prefixes: %w", err)
	}

	glog.V(6).Info("IPv6 prefix deduplication completed")
	return nil
}

// clearExistingClassifiedPrefixes removes existing classified prefixes for reprocessing
func (bdp *BGPDeduplicationProcessor) clearExistingClassifiedPrefixes(ctx context.Context, isIPv4 bool) error {
	var collection string
	if isIPv4 {
		collection = bdp.db.config.BGPPrefixV4
	} else {
		collection = bdp.db.config.BGPPrefixV6
	}

	// Remove all documents from the collection
	query := fmt.Sprintf("FOR doc IN %s REMOVE doc IN %s", collection, collection)
	cursor, err := bdp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to clear %s collection: %w", collection, err)
	}
	defer cursor.Close()

	return nil
}

// processEBGPPrivateIPv4Prefixes processes eBGP private ASN IPv4 prefixes
func (bdp *BGPDeduplicationProcessor) processEBGPPrivateIPv4Prefixes(ctx context.Context) error {
	glog.V(7).Info("Processing eBGP private ASN IPv4 prefixes...")

	// Based on original ipv4-graph logic
	query := `
		FOR u IN unicast_prefix_v4 
		FILTER u.peer_asn IN 64512..65535 
		FILTER u.origin_as IN 64512..65535 
		FILTER u.prefix_len < 30 
		FILTER u.base_attrs.as_path_count == 1 
		FOR p IN peer 
		FILTER u.peer_ip == p.remote_ip 
		INSERT { 
			_key: CONCAT_SEPARATOR("_", u.prefix, u.prefix_len), 
			prefix: u.prefix, 
			prefix_len: u.prefix_len, 
			origin_as: u.origin_as, 
			peer_asn: u.peer_asn,
			nexthop: u.nexthop, 
			peer_ip: u.peer_ip, 
			remote_ip: p.remote_ip, 
			router_id: p.remote_bgp_id,
			prefix_type: "ebgp_private",
		} 
		INTO ` + bdp.db.config.BGPPrefixV4 + ` 
		OPTIONS { ignoreErrors: true }`

	cursor, err := bdp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to process eBGP private IPv4 prefixes: %w", err)
	}
	defer cursor.Close()

	glog.V(7).Info("eBGP private ASN IPv4 prefixes processed")
	return nil
}

// processInternetIPv4Prefixes processes Internet (public ASN) IPv4 prefixes
func (bdp *BGPDeduplicationProcessor) processInternetIPv4Prefixes(ctx context.Context) error {
	glog.V(7).Info("Processing Internet IPv4 prefixes...")

	// Based on original ipv4-graph logic
	// Note: Original used 'u.remote_asn != u.origin_as' but remote_asn doesn't exist in unicast_prefix
	// This effectively made it a no-op, capturing all prefixes including direct announcements
	// Removed the peer_asn != origin_as filter to match original behavior and capture direct announcements
	query := `
		FOR u IN unicast_prefix_v4 
		LET internal_asns = (FOR l IN igp_node RETURN l.peer_asn) 
		FILTER u.peer_asn NOT IN internal_asns 
		FILTER u.peer_asn NOT IN 64512..65535 
		FILTER u.origin_as NOT IN 64512..65535 
		FILTER u.prefix_len < 30 
		INSERT { 
			_key: CONCAT_SEPARATOR("_", u.prefix, u.prefix_len), 
			prefix: u.prefix, 
			prefix_len: u.prefix_len, 
			origin_as: u.origin_as, 
			peer_asn: u.peer_asn,
			nexthop: u.nexthop,
			prefix_type: "ebgp_public",
		} 
		INTO ` + bdp.db.config.BGPPrefixV4 + ` 
		OPTIONS { ignoreErrors: true }`

	cursor, err := bdp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to process Internet IPv4 prefixes: %w", err)
	}
	defer cursor.Close()

	glog.V(7).Info("Internet IPv4 prefixes processed")
	return nil
}

// processIBGPIPv4Prefixes processes iBGP IPv4 prefixes
func (bdp *BGPDeduplicationProcessor) processIBGPIPv4Prefixes(ctx context.Context) error {
	glog.V(7).Info("Processing iBGP IPv4 prefixes...")

	// Based on original ipv4-graph logic - iBGP goes last to sort out externally originated prefixes
	query := `
		FOR u IN unicast_prefix_v4 
		FILTER u.prefix_len < 30 
		FILTER u.base_attrs.local_pref != null 
		FILTER u.base_attrs.as_path_count == null 
		FOR p IN peer 
		FILTER u.peer_ip == p.remote_ip 
		INSERT { 
			_key: CONCAT_SEPARATOR("_", u.prefix, u.prefix_len), 
			prefix: u.prefix, 
			prefix_len: u.prefix_len, 
			peer_asn: u.peer_asn,
			nexthop: u.nexthop, 
			router_id: p.remote_bgp_id, 
			asn: u.peer_asn, 
			local_pref: u.base_attrs.local_pref,
			prefix_type: "ibgp",
		} 
		INTO ` + bdp.db.config.BGPPrefixV4 + ` 
		OPTIONS { ignoreErrors: true }`

	cursor, err := bdp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to process iBGP IPv4 prefixes: %w", err)
	}
	defer cursor.Close()

	glog.V(7).Info("iBGP IPv4 prefixes processed")
	return nil
}

// processEBGPPrivateIPv6Prefixes processes eBGP private ASN IPv6 prefixes (2-byte)
func (bdp *BGPDeduplicationProcessor) processEBGPPrivateIPv6Prefixes(ctx context.Context) error {
	glog.V(7).Info("Processing eBGP private ASN IPv6 prefixes...")

	query := `
		FOR u IN unicast_prefix_v6 
		FILTER u.peer_asn IN 64512..65535 
		FILTER u.origin_as IN 64512..65535 
		FILTER u.prefix_len < 96 
		FILTER u.base_attrs.as_path_count == 1 
		FOR p IN peer 
		FILTER u.peer_ip == p.remote_ip 
		INSERT { 
			_key: CONCAT_SEPARATOR("_", u.prefix, u.prefix_len), 
			prefix: u.prefix, 
			prefix_len: u.prefix_len, 
			origin_as: u.origin_as, 
			peer_asn: u.peer_asn,
			nexthop: u.nexthop, 
			peer_ip: u.peer_ip, 
			remote_ip: p.remote_ip, 
			router_id: p.remote_bgp_id,
			prefix_type: "ebgp_private",
		} 
		INTO ` + bdp.db.config.BGPPrefixV6 + ` 
		OPTIONS { ignoreErrors: true }`

	cursor, err := bdp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to process eBGP private IPv6 prefixes: %w", err)
	}
	defer cursor.Close()

	glog.V(7).Info("eBGP private ASN IPv6 prefixes processed")
	return nil
}

// processEBGP4BytePrivateIPv6Prefixes processes eBGP 4-byte private ASN IPv6 prefixes
func (bdp *BGPDeduplicationProcessor) processEBGP4BytePrivateIPv6Prefixes(ctx context.Context) error {
	glog.V(7).Info("Processing eBGP 4-byte private ASN IPv6 prefixes...")

	query := `
		FOR u IN unicast_prefix_v6 
		FILTER u.peer_asn IN 4200000000..4294967294 
		FILTER u.prefix_len < 96 
		FILTER u.base_attrs.as_path_count == 1 
		FOR p IN peer 
		FILTER u.peer_ip == p.remote_ip 
		INSERT { 
			_key: CONCAT_SEPARATOR("_", u.prefix, u.prefix_len), 
			prefix: u.prefix, 
			prefix_len: u.prefix_len, 
			origin_as: u.origin_as < 0 ? u.origin_as + 4294967296 : u.origin_as, 
			peer_asn: u.peer_asn,
			nexthop: u.nexthop, 
			peer_ip: u.peer_ip, 
			remote_ip: p.remote_ip, 
			router_id: p.remote_bgp_id,
			prefix_type: "ebgp_private_4byte",
		} 
		INTO ` + bdp.db.config.BGPPrefixV6 + ` 
		OPTIONS { ignoreErrors: true }`

	cursor, err := bdp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to process eBGP 4-byte private IPv6 prefixes: %w", err)
	}
	defer cursor.Close()

	glog.V(7).Info("eBGP 4-byte private ASN IPv6 prefixes processed")
	return nil
}

// processInternetIPv6Prefixes processes Internet (public ASN) IPv6 prefixes
func (bdp *BGPDeduplicationProcessor) processInternetIPv6Prefixes(ctx context.Context) error {
	glog.V(7).Info("Processing Internet IPv6 prefixes...")

	// Based on original ipv6-graph logic
	// Note: Original used 'u.remote_asn != u.origin_as' but remote_asn doesn't exist in unicast_prefix
	// This effectively made it a no-op, capturing all prefixes including direct announcements
	// Removed the peer_asn != origin_as filter to match original behavior and capture direct announcements
	query := `
		FOR u IN unicast_prefix_v6 
		LET internal_asns = (FOR l IN igp_node RETURN l.peer_asn) 
		FILTER u.peer_asn NOT IN internal_asns 
		FILTER u.peer_asn NOT IN 64512..65535 
		FILTER u.peer_asn NOT IN 4200000000..4294967294 
		FILTER u.origin_as NOT IN 64512..65535 
		FILTER u.prefix_len < 96 
		INSERT { 
			_key: CONCAT_SEPARATOR("_", u.prefix, u.prefix_len), 
			prefix: u.prefix, 
			prefix_len: u.prefix_len, 
			origin_as: u.origin_as, 
			peer_asn: u.peer_asn,
			nexthop: u.nexthop,
			prefix_type: "ebgp_public",
		} 
		INTO ` + bdp.db.config.BGPPrefixV6 + ` 
		OPTIONS { ignoreErrors: true }`

	cursor, err := bdp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to process Internet IPv6 prefixes: %w", err)
	}
	defer cursor.Close()

	glog.V(7).Info("Internet IPv6 prefixes processed")
	return nil
}

// processIBGPIPv6Prefixes processes iBGP IPv6 prefixes
func (bdp *BGPDeduplicationProcessor) processIBGPIPv6Prefixes(ctx context.Context) error {
	glog.V(7).Info("Processing iBGP IPv6 prefixes...")

	query := `
		FOR u IN unicast_prefix_v6 
		FILTER u.prefix_len < 96 
		FILTER u.base_attrs.local_pref != null 
		FILTER u.base_attrs.as_path_count == null 
		FOR p IN peer 
		FILTER u.peer_ip == p.remote_ip 
		INSERT { 
			_key: CONCAT_SEPARATOR("_", u.prefix, u.prefix_len), 
			prefix: u.prefix, 
			prefix_len: u.prefix_len, 
			peer_asn: u.peer_asn,
			nexthop: u.nexthop, 
			router_id: p.remote_bgp_id, 
			asn: u.peer_asn, 
			local_pref: u.base_attrs.local_pref,
			prefix_type: "ibgp",
		} 
		INTO ` + bdp.db.config.BGPPrefixV6 + ` 
		OPTIONS { ignoreErrors: true }`

	cursor, err := bdp.db.db.Query(ctx, query, nil)
	if err != nil {
		return fmt.Errorf("failed to process iBGP IPv6 prefixes: %w", err)
	}
	defer cursor.Close()

	glog.V(7).Info("iBGP IPv6 prefixes processed")
	return nil
}
