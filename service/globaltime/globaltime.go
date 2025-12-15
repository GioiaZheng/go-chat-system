package globaltime

import "time"

// FixedTime holds a deterministic timestamp used to make time-dependent
// operations predictable in tests. When it is the zero value, realtime is used.
var FixedTime time.Time

// Now returns FixedTime when it has been set, otherwise it mirrors time.Now.
// This helper keeps production logic and tests aligned on a single entrypoint.
func Now() time.Time {
	if FixedTime.After(time.Time{}) {
		return FixedTime
	}
	return time.Now()
}

// Since calculates the elapsed duration from the provided timestamp using Now.
func Since(tm time.Time) time.Duration {
	return Now().Sub(tm)
}
