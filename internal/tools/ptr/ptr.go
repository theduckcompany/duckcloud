package ptr

// To returns a pointer from any types
func To[K any](input K) *K {
	return &input
}
