package shellrun

// TruncateBytes returns a possibly truncated view of b with details.
// If max <= 0 or len(b) <= max -> return b, false, len(b)
// mode: "head" keeps first max bytes; "tail" keeps last max bytes (default head on unknown).
func TruncateBytes(b []byte, maxBytes int, mode string) (seg []byte, truncated bool, total int) {
	total = len(b)
	if maxBytes <= 0 || len(b) <= maxBytes {
		return b, false, total
	}
	if mode == "tail" {
		return b[len(b)-maxBytes:], true, total
	}
	return b[:maxBytes], true, total
}
