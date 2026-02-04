package utils

func SliceToMap(slice []string) map[string]bool {
	m := make(map[string]bool)
	for _, item := range slice {
		m[item] = true
	}
	return m
}
