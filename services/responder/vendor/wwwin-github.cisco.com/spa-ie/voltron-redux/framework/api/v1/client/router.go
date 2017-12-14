/*
 * Voltron Framework
 *
 * This outlines version 1 of the voltron framework API.
 *
 * OpenAPI spec version: 1.0
 *
 * Generated by: https://github.com/swagger-api/swagger-codegen.git
 */

package client

type Router struct {
	Key string `json:"_key,omitempty"`

	Name string `json:"_name,omitempty"`

	RouterIP string `json:"RouterIP,omitempty"`

	BGPID string `json:"BGPID,omitempty"`

	IsLocal bool `json:"IsLocal,omitempty"`

	ASN string `json:"ASN,omitempty"`
}
