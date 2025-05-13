package api

// Close should close everything opened in the lifecycle of the `_router`; for example, background goroutines.
func (rt *_router) Close() error {
	if rt.db != nil {
		if err := rt.db.Close(); err != nil {
			return err
		}
	}
	return nil
}
