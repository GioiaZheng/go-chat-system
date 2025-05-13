package database

// Helper function to convert UUID to int64 if needed
func int64FromUUID(uuidStr string) int64 {
	hash := int64(0)
	for _, ch := range uuidStr {
		hash = hash*31 + int64(ch)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}
