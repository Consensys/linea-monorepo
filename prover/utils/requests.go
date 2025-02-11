package utils

// CombineRequests combines multiple request files (slices) into a single request slice
func CombineRequests(requests ...[]string) []string {
	// Calculate the total length of the combined slice
	totalLength := 0
	for _, request := range requests {
		totalLength += len(request)
	}

	// Preallocate the combined slice with the total length
	combined := make([]string, totalLength)

	// Copy each slice into the combined slice
	currentIndex := 0
	for _, request := range requests {
		copy(combined[currentIndex:], request)
		currentIndex += len(request)
	}

	return combined
}
