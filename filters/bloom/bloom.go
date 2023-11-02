// Documentation:
// https://pkg.go.dev/github.com/bits-and-blooms/bloom/v3

package main

import (
	"fmt"

	"github.com/bits-and-blooms/bloom/v3"
)

func main() {

	// Create a new Bloom Filter with 1,000,000 items and 0.01% false positive rate
	bf := bloom.NewWithEstimates(1000000, 0.01)
	bf.Add([]byte("Hello, world!"))

	// Test for existence (false positive)
	if bf.Test([]byte("Hello, orld!")) {
		fmt.Println("Exists!")
	}
}
