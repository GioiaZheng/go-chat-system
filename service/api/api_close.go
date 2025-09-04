package api

// Close implements a no-op shutdown hook for the API router.
// It allows cmd/webapi/main.go to call apirouter.Close() without errors.
// If you later need to release resources (e.g., background workers), add them here.
func (rt *_router) Close() error {
	// No resources to clean up in this assignment.
	return nil
}
