package nextcloud

import "encoding/xml"

// parseXML is a helper for parsing DAV XML responses with custom namespace handling.
func parseXML(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}
