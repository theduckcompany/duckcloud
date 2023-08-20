package ptr

// To returns a pointer from any types
func To[K any](input K) *K {
	return &input
}

// From returns the value from any pointer type
func From[K any](input *K) K {
	if input == nil {
		var defaultInput K
		return defaultInput
	}
	return *input
}
