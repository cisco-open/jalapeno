package arangodb

import (
	"strconv"
)

func checkASNLocation(asn int32) bool {
	var isInternalASN bool = false
	if ((int(asn) >= 64512) && (int(asn) <= 65535)) || ((int(asn) >= 4200000000) && (int(asn) <= 4294967294)) {
		isInternalASN = true
	}
	return isInternalASN
}

// Calculates sid value using initial SRGB label and sid-index
func calculateSID(srBeginningLabel int, sidIndex string) int {
	sidIndexVal, _ := strconv.ParseInt(sidIndex, 10, 0)
	sidVal := int(srBeginningLabel) + int(sidIndexVal)
	return sidVal
}
