package main

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/taurusgroup/multi-party-sig/internal/test"
	"github.com/taurusgroup/multi-party-sig/pkg/ecdsa"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/pool"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
	"github.com/taurusgroup/multi-party-sig/pkg/taproot"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp"
	"github.com/taurusgroup/multi-party-sig/protocols/example"
	"github.com/taurusgroup/multi-party-sig/protocols/frost"

	// cuckoo "mpc-wallet-multiple-filters/filters/cuckoo"
	"github.com/bits-and-blooms/bloom/v3"

	"golang.org/x/crypto/sha3"
)

var lookupAddress string
var message []byte
var mu sync.Mutex

// var filter = cuckoo.NewCuckooFilter(10, 0.001)
var filter = bloom.NewWithEstimates(1000000, 0.01)
var (
	// The number of buckets
	globalNumBuckets uint
	// Mutex to handle concurrent access to the global variable
	bucketMutex sync.Mutex
)

// Add crypto address to the bloom filter.
func addCryptoAddressToFilter(cryptoAddress string) {
	filter.Add([]byte(cryptoAddress))
}

// AppendFilterToMessage appends the bloom filter to the given message.
func AppendFilterToMessage(m []byte, filter *bloom.BloomFilter) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	// Serialize the filter.
	err := encoder.Encode(filter)
	if err != nil {
		return nil, err
	}

	// Append the serialized filter to the message.
	return append(m, buffer.Bytes()...), nil
}

// ExtractFilterFromMessage extracts the bloom filter from the appended message.
func ExtractFilterFromMessage(appendedMessage []byte) ([]byte, *bloom.BloomFilter, error) {
	buffer := bytes.NewBuffer(appendedMessage)
	decoder := gob.NewDecoder(buffer)
	var filter bloom.BloomFilter

	// Try decoding from the end until it succeeds.
	for i := len(appendedMessage); i >= 0; i-- {
		buffer = bytes.NewBuffer(appendedMessage[i:])
		decoder = gob.NewDecoder(buffer)
		err := decoder.Decode(&filter)
		if err == nil {
			// Decoding succeeded, return the original message and the decoded filter.
			return appendedMessage[:i], &filter, nil
		}
	}
	return nil, nil, errors.New("could not decode the bloom filter from the message")
}

func checkCryptoAddressInFilter(address string) bool {
	mu.Lock()
	defer mu.Unlock()
	return filter.Test([]byte(address))
}

// Given a public key, this function will compute the Ethereum address
func PublicKeyToCryptoAddress(pub []byte) string {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(pub[1:]) // remove EC prefix 04
	return fmt.Sprintf("0x%x", hash.Sum(nil)[12:])
}

func ConvertToCryptoAddress(s curve.Scalar) (string, error) {
	// Convert the scalar to bytes
	data, err := s.MarshalBinary()
	if err != nil {
		return "", err
	}

	// For crypto address generation, we treat the scalar bytes as a public key.
	// This might not be accurate in a real-world scenario. But just to demonstrate:
	return PublicKeyToCryptoAddress(data), nil
}
func XOR(id party.ID, ids party.IDSlice, n *test.Network) error {
	h, err := protocol.NewMultiHandler(example.StartXOR(id, ids), nil)
	if err != nil {
		return err
	}
	test.HandlerLoop(id, h, n)
	_, err = h.Result()
	if err != nil {
		return err
	}
	return nil
}

func CMPKeygen(id party.ID, ids party.IDSlice, threshold int, n *test.Network, pl *pool.Pool) (*cmp.Config, error) {
	h, err := protocol.NewMultiHandler(cmp.Keygen(curve.Secp256k1{}, id, ids, threshold, pl), nil)
	if err != nil {
		return nil, err
	}
	test.HandlerLoop(id, h, n)
	r, err := h.Result()
	if err != nil {
		return nil, err
	}

	return r.(*cmp.Config), nil
}

func CMPRefresh(c *cmp.Config, n *test.Network, pl *pool.Pool) (*cmp.Config, error) {
	hRefresh, err := protocol.NewMultiHandler(cmp.Refresh(c, pl), nil)
	if err != nil {
		return nil, err
	}
	test.HandlerLoop(c.ID, hRefresh, n)

	r, err := hRefresh.Result()
	if err != nil {
		return nil, err
	}

	return r.(*cmp.Config), nil
}

func CMPSign(c *cmp.Config, m []byte, signers party.IDSlice, n *test.Network, pl *pool.Pool) error {
	h, err := protocol.NewMultiHandler(cmp.Sign(c, signers, m, pl), nil)
	if err != nil {
		return err
	}
	test.HandlerLoop(c.ID, h, n)

	cryptoAddress, err := ConvertToCryptoAddress(c.ECDSA)
	lookupAddress = cryptoAddress
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	signResult, err := h.Result()
	if err != nil {
		return err
	}
	signature := signResult.(*ecdsa.Signature)

	// Start timing the verification
	// startTime := time.Now()

	if !signature.Verify(c.PublicPoint(), m) {
		return errors.New("failed to verify cmp signature")
	}

	// End timing and calculate elapsed time in microseconds
	// elapsedTime := time.Since(startTime).Microseconds()
	// fmt.Printf("Time taken to verify: %dµs\n", elapsedTime)

	// Update cuckoo filter
	addCryptoAddressToFilter(cryptoAddress)
	// if checkCryptoAddressInFilter(cryptoAddress) {
	// 	fmt.Println("Crypto address is in the filter")
	// } else {
	// 	fmt.Println("Crypto address is not in the filter")
	// }
	appendedMessage, err := AppendFilterToMessage(m, filter)
	if err != nil {
		fmt.Println("Error appending filter to message:", err)
		return err
	}
	message = appendedMessage
	// extractedMessage, extractedFilter, err := ExtractFilterFromMessage(appendedMessage)
	// if err != nil {
	// 	fmt.Println("Error extracting filter from appended message:", err, extractedMessage)
	// 	return err
	// }
	// test the cryptoaddress presence in the exrtracted filter
	// if extractedFilter.Test([]byte(cryptoAddress)) {
	// 	fmt.Println("Crypto address is in the extracted filter")
	// } else {
	// 	fmt.Println("Crypto address is not in the extracted filter")
	// }

	return nil
}

func CMPPreSign(c *cmp.Config, signers party.IDSlice, n *test.Network, pl *pool.Pool) (*ecdsa.PreSignature, error) {
	h, err := protocol.NewMultiHandler(cmp.Presign(c, signers, pl), nil)
	if err != nil {
		return nil, err
	}

	test.HandlerLoop(c.ID, h, n)

	signResult, err := h.Result()
	if err != nil {
		return nil, err
	}

	preSignature := signResult.(*ecdsa.PreSignature)
	if err = preSignature.Validate(); err != nil {
		return nil, errors.New("failed to verify cmp presignature")
	}
	return preSignature, nil
}

func CMPPreSignOnline(c *cmp.Config, preSignature *ecdsa.PreSignature, m []byte, n *test.Network, pl *pool.Pool) error {
	h, err := protocol.NewMultiHandler(cmp.PresignOnline(c, preSignature, m, pl), nil)
	if err != nil {
		return err
	}
	test.HandlerLoop(c.ID, h, n)
	// print singers and their ids
	// fmt.Println("signers: ", preSignature.SignerIDs())

	signResult, err := h.Result()
	if err != nil {
		return err
	}
	signature := signResult.(*ecdsa.Signature)
	if !signature.Verify(c.PublicPoint(), m) {
		return errors.New("failed to verify cmp signature")
	}
	return nil
}

func FrostKeygen(id party.ID, ids party.IDSlice, threshold int, n *test.Network) (*frost.Config, error) {
	h, err := protocol.NewMultiHandler(frost.Keygen(curve.Secp256k1{}, id, ids, threshold), nil)
	if err != nil {
		return nil, err
	}
	test.HandlerLoop(id, h, n)
	r, err := h.Result()
	if err != nil {
		return nil, err
	}

	return r.(*frost.Config), nil
}

func FrostSign(c *frost.Config, id party.ID, m []byte, signers party.IDSlice, n *test.Network) error {
	h, err := protocol.NewMultiHandler(frost.Sign(c, signers, m), nil)
	if err != nil {
		return err
	}
	test.HandlerLoop(id, h, n)
	r, err := h.Result()
	if err != nil {
		return err
	}

	signature := r.(frost.Signature)
	if !signature.Verify(c.PublicKey, m) {
		return errors.New("failed to verify frost signature")
	}
	return nil
}

func FrostKeygenTaproot(id party.ID, ids party.IDSlice, threshold int, n *test.Network) (*frost.TaprootConfig, error) {
	h, err := protocol.NewMultiHandler(frost.KeygenTaproot(id, ids, threshold), nil)
	if err != nil {
		return nil, err
	}
	test.HandlerLoop(id, h, n)
	r, err := h.Result()
	if err != nil {
		return nil, err
	}

	return r.(*frost.TaprootConfig), nil
}
func FrostSignTaproot(c *frost.TaprootConfig, id party.ID, m []byte, signers party.IDSlice, n *test.Network) error {
	h, err := protocol.NewMultiHandler(frost.SignTaproot(c, signers, m), nil)
	if err != nil {
		return err
	}
	test.HandlerLoop(id, h, n)
	r, err := h.Result()
	if err != nil {
		return err
	}

	signature := r.(taproot.Signature)
	if !c.PublicKey.Verify(signature, m) {
		return errors.New("failed to verify frost signature")
	}
	return nil
}

func All(id party.ID, ids party.IDSlice, threshold int, message []byte, n *test.Network, wg *sync.WaitGroup, pl *pool.Pool) error {
	defer wg.Done()

	// XOR
	err := XOR(id, ids, n)
	if err != nil {
		return err
	}

	// CMP KEYGEN
	keygenConfig, err := CMPKeygen(id, ids, threshold, n, pl)
	if err != nil {
		return err
	}

	// CMP REFRESH
	refreshConfig, err := CMPRefresh(keygenConfig, n, pl)
	if err != nil {
		return err
	}
	// print ids
	// fmt.Println(refreshConfig.ID, refreshConfig.ECDSA)
	// // Usage
	// address, err := ConvertToCryptoAddress(refreshConfig.ECDSA)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return err
	// }
	// fmt.Println("Crypto address:", address)
	// fmt.Println("ID: ", refreshConfig.ID)

	// FROST KEYGEN
	// frostResult, err := FrostKeygen(id, ids, threshold, n)
	// if err != nil {
	// 	return err
	// }

	// // FROST KEYGEN TAPROOT
	// frostResultTaproot, err := FrostKeygenTaproot(id, ids, threshold, n)
	// if err != nil {
	// 	return err
	// }

	signers := ids[:threshold+1]
	if !signers.Contains(id) {
		n.Quit(id)
		return nil
	}
	// fmt.Println("signers: ", signers)
	// print pool

	// CMP SIGN
	err = CMPSign(refreshConfig, message, signers, n, pl)
	if err != nil {
		return err
	}

	// CMP PRESIGN
	preSignature, err := CMPPreSign(refreshConfig, signers, n, pl)
	if err != nil {
		return err
	}

	// CMP PRESIGN ONLINE
	err = CMPPreSignOnline(refreshConfig, preSignature, message, n, pl)
	if err != nil {
		return err
	}

	// FROST SIGN
	// err = FrostSign(frostResult, id, message, signers, n)
	// if err != nil {
	// 	return err
	// }

	// // FROST SIGN TAPROOT
	// err = FrostSignTaproot(frostResultTaproot, id, message, signers, n)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func main() {
	partySets := []party.IDSlice{
		// {"a", "b", "c"},
		// {"a", "b", "c", "d"},
		// {"a", "b", "c", "d", "e"},
		// {"a", "b", "c", "d", "e", "f"},
		{"a", "b", "c", "d", "e", "f", "g"},
	}
	threshold := 2
	messageToSign := []byte("hello1")
	// lookupAddress := []byte("someCryptoAddress")

	// Prepare the CSV file for writing
	file, err := os.Create("results.csv")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"Party Setup", "Iteration", "Extract Time (µs)", "Lookup Time (µs)", "Combined Time (µs)", "Result"})

	for _, ids := range partySets {
		// var totalExtractTime, totalLookupTime

		for i := 0; i < 10; i++ {
			filter = bloom.NewWithEstimates(1000000, 0.01)
			net := test.NewNetwork(ids)

			var wg sync.WaitGroup
			for _, id := range ids {
				wg.Add(1)
				go func(id party.ID) {
					pl := pool.NewPool(0)
					defer pl.TearDown()
					if err := All(id, ids, threshold, messageToSign, net, &wg, pl); err != nil {
						fmt.Println(err)
					}
				}(id)
			}
			wg.Wait()

			// Measure Extract time
			startExtract := time.Now()
			extractedMessage, extractedFilter, err := ExtractFilterFromMessage(message)
			// print extcated message
			endExtract := time.Now()

			fmt.Println("Extracted message and filter: ", extractedMessage, extractedFilter)

			if err != nil {
				fmt.Println("Error extracting filter from appended message:", err)
				return
			}

			// Measure Lookup time
			startLookup := time.Now()
			lookupResult := filter.Test([]byte(lookupAddress))
			endLookup := time.Now()

			// Measure Combined time

			extractTimeMicro := endExtract.Sub(startExtract).Microseconds()
			lookupTimeMicro := endLookup.Sub(startLookup).Microseconds()
			combinedTimeMicro := extractTimeMicro + lookupTimeMicro

			// Write results to CSV
			writer.Write([]string{
				fmt.Sprint(ids),
				fmt.Sprint(i + 1),
				fmt.Sprintf("%.2f", float64(extractTimeMicro)),
				fmt.Sprintf("%.2f", float64(lookupTimeMicro)),
				fmt.Sprintf("%.2f", float64(combinedTimeMicro)),
				fmt.Sprintf("%v", lookupResult),
			})
		}
		// Print average times to console
		// fmt.Println("Party Setup:", ids)
		// fmt.Println("Average Extract Time (µs):", totalExtractTime.Microseconds()/10)
		// fmt.Println("Average Lookup Time (µs):", totalLookupTime.Microseconds()/10)
		// fmt.Println("Average Combined Time (µs):", totalCombinedTime.Microseconds()/10)

	}
}
