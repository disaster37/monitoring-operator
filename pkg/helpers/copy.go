package helpers

// CopyMapString permit to copy map[string]string
// It usefull to copy labels or annotations from parent resources
func CopyMapString(src map[string]string) (dst map[string]string) {
	if src == nil {
		return nil
	}

	dst = map[string]string{}
	for key, value := range src {
		dst[key] = value
	}

	return dst
}
