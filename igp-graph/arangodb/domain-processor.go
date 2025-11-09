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

	driver "github.com/arangodb/go-driver"
	"github.com/golang/glog"
)

// Domain represents an IGP domain in the network
type Domain struct {
	Key        string `json:"_key,omitempty"`
	ID         string `json:"_id,omitempty"`
	Rev        string `json:"_rev,omitempty"`
	ASN        uint32 `json:"asn"`
	ProtocolID int    `json:"protocol_id"`
	DomainID   int64  `json:"domain_id"`
	Protocol   string `json:"protocol,omitempty"`
}

// ensureIGPDomain creates or updates an IGP domain entry
func (a *arangoDB) ensureIGPDomain(ctx context.Context, node map[string]interface{}) error {
	// Extract domain information from node
	protocolID, ok1 := node["protocol_id"]
	domainID, ok2 := node["domain_id"]
	asn, ok3 := node["asn"]
	protocol, _ := node["protocol"].(string)

	if !ok1 || !ok2 || !ok3 {
		return fmt.Errorf("missing required domain fields in node data")
	}

	// Create domain key: ProtocolID_DomainID_ASN
	domainKey := fmt.Sprintf("%v_%v_%v", protocolID, domainID, asn)

	// Create domain document
	domainDoc := Domain{
		Key:        domainKey,
		ASN:        getUint32(asn),
		ProtocolID: getInt(protocolID),
		DomainID:   getInt64(domainID),
		Protocol:   protocol,
	}

	// Try to create the domain document
	_, err := a.igpDomain.CreateDocument(ctx, domainDoc)
	if err != nil {
		if driver.IsConflict(err) {
			// Domain already exists, which is fine
			glog.V(8).Infof("IGP domain %s already exists", domainKey)
			return nil
		}
		return fmt.Errorf("failed to create IGP domain %s: %w", domainKey, err)
	}

	glog.V(6).Infof("Created IGP domain: %s (%s)", domainKey, protocol)
	return nil
}

// Helper functions for type conversion
func getUint32(v interface{}) uint32 {
	switch val := v.(type) {
	case uint32:
		return val
	case float64:
		return uint32(val)
	case int:
		return uint32(val)
	default:
		return 0
	}
}

func getInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case uint32:
		return int(val)
	default:
		return 0
	}
}

func getInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	case int:
		return int64(val)
	case uint32:
		return int64(val)
	default:
		return 0
	}
}
