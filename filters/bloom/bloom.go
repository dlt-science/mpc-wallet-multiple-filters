// Documentation:
// https://pkg.go.dev/github.com/bits-and-blooms/bloom/v3

package main

import (
	"fmt"

	"github.com/bits-and-blooms/bloom/v3"
)

func main() {

	// Create a new Bloom Filter with 1,000,000 items and 0.01% false positive rate
	bf := bloom.NewWithEstimates(3, 0.01)
	bf.Add([]byte("Hello, world!"))

	bf.K()
	// print the value fo k
	fmt.Println(bloom.EstimateParameters(3, 0.01))
	// Test for existence (false positive)
	if bf.Test([]byte("Hello, orld!")) {
		fmt.Println("Exists!")
	}
}
