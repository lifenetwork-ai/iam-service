package utils

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/exp/rand"
)

// SelectRandomSubset selects a random subset of elements from a slice.
// The subset size must be an odd number (e.g., 3, 5, 7, ...).
func SelectRandomSubset[T any](items []T, subsetSize int) ([]T, error) {
	// Ensure the subset size is odd
	if subsetSize%2 == 0 {
		return nil, errors.New("subset size must be an odd number")
	}

	// If there are fewer items than the requested subset size, return all items
	if len(items) <= subsetSize {
		return items, nil
	}

	// Seed the random number generator
	rand.Seed(uint64(time.Now().UnixNano()))

	// Shuffle the items slice
	rand.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})

	// Return the first `subsetSize` elements
	return items[:subsetSize], nil
}

// GenerateOTP generates a random 6-digit one-time password.
func GenerateOTP() string {
	// Seed the random number generator
	rand.Seed(uint64(time.Now().UnixNano()))

	// Generate a random 6-digit number
	return fmt.Sprintf("%06d", rand.Intn(1_000_000))
}
