package utils

func InStringSlice(needle string, haystack []string) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}
	return false
}
