package main

import (
	"bytes"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Create a new CSV file
	file, err := os.Create("BitcoinBloomResults.csv")
	if err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}
	defer file.Close()

	// Create a new CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header row to the CSV file
	err = writer.Write([]string{"partySet", "filterSize", "transactionSize", "transactionFeeBTC", "transactionFeeSatoshis"})
	if err != nil {
		log.Fatalf("Failed to write to CSV file: %v", err)
	}

	// Define the party sets to test
	partySets := [][]int{
		{1, 1},
		{2, 2},
		{2, 3},
		{3, 4},
		{4, 5},
		{5, 10},
	}

	// Fetch the current fee rate from the mempool.space API to calculate the transaction fee
	feeRate := getCurrentFeeRate()

	for _, partySet := range partySets {
		// Generate a set of private keys
		privateKeys := generatePrivateKeys(partySet[0])

		// Generate a set of public keys
		publicKeys := generatePublicKeys(privateKeys)

		// Generate a Cuckoo filter for the private keys
		// Generating a filter for total accepted items for a set. E.g. filter of 4 which would have 3 items
		serializedCuckooFilter := getBloomFilter(publicKeys, uint(partySet[1]), 0.0001)

		// Convert the serializedCuckooFilter to []byte if it's not already in that format
		serializedCuckooFilterBytes := []byte(serializedCuckooFilter)
		println("The size of the serializedCuckooFilterBytes is: ", len(serializedCuckooFilterBytes), " bytes")

		// Create a new script builder
		builder := txscript.NewScriptBuilder()

		// Add the serializedCuckooFilter to the script
		builder.AddData(serializedCuckooFilterBytes)

		// Get the final script
		scriptSig, err := builder.Script()
		if err != nil {
			log.Fatalf("Failed to create script: %v", err)
		}

		println("The size of the scriptSig is: ", len(scriptSig), " bytes")

		// Create a new message transaction
		msgTx := wire.NewMsgTx(wire.TxVersion)

		// Create a new transaction using the serializedCuckooFilterBytes as a ScriptPubKey
		// Sending 0.0001 BTC to the address, which equals 10,000 satoshis
		txOut := wire.NewTxOut(10000, scriptSig)

		// Add the output to the transaction
		msgTx.AddTxOut(txOut)

		// Format the addresses to send the transaction from and to
		privKeyBytes, err := hex.DecodeString(os.Getenv("PRIVATE_KEY"))
		if err != nil {
			log.Fatalf("Failed to decode private key: %v", err)
		}
		privateKey, _ := btcec.PrivKeyFromBytes(privKeyBytes)

		// Create a new transaction input
		prevTxID := wire.NewOutPoint(&chainhash.Hash{}, ^uint32(0)) // replace with your previous transaction ID and index
		txIn := wire.NewTxIn(prevTxID, nil, nil)

		// Add the input to the transaction
		msgTx.AddTxIn(txIn)

		// Sign the transaction
		signature, err := txscript.SignatureScript(msgTx, 0, scriptSig, txscript.SigHashAll, privateKey, false)
		if err != nil {
			log.Fatalf("Failed to sign transaction: %v", err)
		}

		// set the SignatureScript on the transaction input.
		msgTx.TxIn[0].SignatureScript = signature

		// Measure the transaction size
		txSize := getTxSize(msgTx)
		println("The transaction size is: ", txSize, " bytes")

		// Measure the transaction fee
		txFeeSatoshis := txSize * feeRate
		println("The transaction fee is: ", txFeeSatoshis, " satoshis")

		//txOutValue := msgTx.TxOut[0].Value
		//println("The transaction output value is: ", txOutValue, " satoshis")

		// Calculate the transaction fee
		//txOutValueInBTC := float64(txOutValue) / 100000000
		//println("The transaction fee is: ", txOutValueInBTC, " BTC")

		txFeeInBTC := float64(txFeeSatoshis) / 100000000
		println("The transaction fee is: ", txFeeInBTC, " BTC")

		// Write the results to the CSV file
		err = writer.Write([]string{
			fmt.Sprintf("%d of %d", partySet[0], partySet[1]),
			strconv.Itoa(len(serializedCuckooFilterBytes)),
			strconv.Itoa(txSize),
			fmt.Sprintf("%.8f", txFeeInBTC),
			strconv.Itoa(int(txFeeSatoshis)),
		})
		if err != nil {
			log.Fatalf("Failed to write to CSV file: %v", err)
		}
	}
}

func generateRandomPrivateKey() string {
	privateKey := make([]byte, 32)
	_, err := rand.Read(privateKey)
	if err != nil {
		log.Fatalf("Failed to generate random private key: %v", err)
	}
	return hex.EncodeToString(privateKey)
}

func generatePrivateKeys(numKeys int) []string {
	privateKeys := make([]string, numKeys)
	for i := 0; i < numKeys; i++ {
		privateKeys[i] = generateRandomPrivateKey()
	}
	return privateKeys
}

func generatePublicKeys(privateKeys []string) []string {
	publicKeys := make([]string, len(privateKeys))
	for i, privateKey := range privateKeys {
		privKeyBytes, _ := hex.DecodeString(privateKey)
		_, pubKey := btcec.PrivKeyFromBytes(privKeyBytes)
		pubKeyBytes := pubKey.SerializeCompressed()
		publicKeys[i] = hex.EncodeToString(pubKeyBytes)
	}
	return publicKeys
}

func getBloomFilter(publicKeys []string, n uint, fp float64) []byte {

	if n == 0 {
		n = 3 // default value
	}
	if fp == 0 {
		fp = 0.01 // default value
	}
	//Create a new Bloom Filter with 1,000,000 items and 0.001% false positive rate
	bf := bloom.NewWithEstimates(n, fp)

	for _, pubKey := range publicKeys {
		pubKeyBytes, _ := hex.DecodeString(pubKey)
		bf.Add(pubKeyBytes)

		if bf.Test(pubKeyBytes) {
			fmt.Println(pubKeyBytes, " Exists!")
		}
	}

	// Serialize the Bloom filter
	// Assuming `b` is your bloom filter
	var buf bytes.Buffer
	_, err := bf.WriteTo(&buf)
	if err != nil {
		log.Fatalf("Error serializing the Bloom filter: %v", err)
	}

	serializedBloomFilter := buf.Bytes()

	return serializedBloomFilter
}

func getTxSize(msgTx *wire.MsgTx) int {
	// Serialize the transaction
	var buf bytes.Buffer
	err := msgTx.Serialize(&buf)
	if err != nil {
		log.Fatalf("Failed to serialize transaction: %v", err)
	}

	// Measure the transaction size
	txSize := len(buf.Bytes())
	fmt.Printf("Transaction size: %d bytes\n", txSize)

	return txSize
}

//type FeeInfo struct {
//	FastestFee  int `json:"fastestFee"`
//	HalfHourFee int `json:"halfHourFee"`
//	HourFee     int `json:"hourFee"`
//}

type FeeInfo struct {
	FastestFee int `json:"fastestFee"`
}

func getCurrentFeeRate() int {
	resp, err := http.Get("https://mempool.space/api/v1/fees/recommended")
	if err != nil {
		log.Fatalf("Failed to fetch fee rates: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var feeInfo FeeInfo
	err = json.Unmarshal(body, &feeInfo)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Return the fastest fee rate
	return feeInfo.FastestFee
}

//func getCurrentFeeRate() int {
//	resp, err := http.Get("https://bitcoinfees.earn.com/api/v1/fees/recommended")
//	if err != nil {
//		log.Fatalf("Failed to fetch fee rates: %v", err)
//	}
//	defer resp.Body.Close()
//
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		log.Fatalf("Failed to read response body: %v", err)
//	}
//
//	// Print out the response body
//	fmt.Println("Response body:", string(body))
//
//	var feeInfo FeeInfo
//	err = json.Unmarshal(body, &feeInfo)
//	if err != nil {
//		log.Fatalf("Failed to unmarshal JSON: %v", err)
//	}
//
//	// Return the fastest fee rate
//	return feeInfo.FastestFee
//}

//
//	// Add the public keys to the filter
//	if pubKey1 != "" {
//		bf.Add([]byte(pubKey1))
//	}
//	if pubKey2 != "" {
//		bf.Add([]byte(pubKey2))
//	}
//	if pubKey3 != "" {
//		bf.Add([]byte(pubKey3))
//	}
//
//	bf.K()
//
//	// print the value fo k
//	fmt.Println(bloom.EstimateParameters(n, fp))
//
//	// Test for existence (false positive)
//	if bf.Test([]byte(pubKey1)) {
//		fmt.Println("pubKey1 Exists!")
//	}
//
//	if bf.Test([]byte(pubKey2)) {
//		fmt.Println("pubKey2 Exists!")
//	}
//
//	if bf.Test([]byte(pubKey3)) {
//		fmt.Println("pubKey3 Exists!")
//	}
//
//	//// Serialize the Bloom filter
//	//serializedBloomFilter, err := bf.MarshalJSON()
//	//if err != nil {
//	//	log.Fatalf("Failed to serialize Bloom filter: %v", err)
//	//}
//
//	// Serialize the Bloom filter
//	// Assuming `b` is your bloom filter
//	var buf bytes.Buffer
//	_, err := bf.WriteTo(&buf)
//	if err != nil {
//		log.Fatalf("Error serializing the Bloom filter: %v", err)
//	}
//
//	serializedBloomFilter := buf.Bytes()
//
//	//	return serializedBloomFilter
//}

//func genPubKeys(PVK_1 string, PVK_2 string, PVK_3 string) (bytes string, bytes2 string, bytes3 string) {
//	// Based on: https://github.com/btcsuite/btcd/blob/master/txscript/example_test.go#L84
//	//PVK_1, err := hex.DecodeString(PVK_1)
//	//if err != nil {
//	//	fmt.Println(err)
//	//	return
//	//}
//	//PVK_2, err := hex.DecodeString(PVK_2)
//	//if err != nil {
//	//	fmt.Println(err)
//	//	return
//	//}
//	//PVK_3, err := hex.DecodeString(PVK_3)
//	//if err != nil {
//	//	fmt.Println(err)
//	//	return
//	//}
//
//	// Generate the public keys from the private keys
//	_, pubKey1 := btcec.PrivKeyFromBytes([]byte(PVK_1))
//	//pubKey1 := privKey1.PubKey()
//
//	_, pubKey2 := btcec.PrivKeyFromBytes([]byte(PVK_2))
//	//pubKey2 := privKey2.PubKey()
//
//	_, pubKey3 := btcec.PrivKeyFromBytes([]byte(PVK_3))
//	//pubKey3 := privKey3.PubKey()
//
//	// Convert the public keys to bytes
//	pubKey1Bytes := pubKey1.SerializeCompressed()
//	pubKey2Bytes := pubKey2.SerializeCompressed()
//	pubKey3Bytes := pubKey3.SerializeCompressed()
//
//	// Convert the bytes to hexadecimal strings
//	pubKey1String := hex.EncodeToString(pubKey1Bytes)
//	pubKey2String := hex.EncodeToString(pubKey2Bytes)
//	pubKey3String := hex.EncodeToString(pubKey3Bytes)
//
//	return pubKey1String, pubKey2String, pubKey3String
//}
//

//func getBloomFilter(pubKey1 string, pubKey2 string, pubKey3 string, n uint, fp float64) []byte {
//	// Check if n and fp are zero, and if so, assign them default values
//	if n == 0 {
//		n = 3 // default value
//	}
//	if fp == 0 {
//		fp = 0.01 // default value
//	}
//
//	// Create a new Bloom Filter with 1,000,000 items and 0.001% false positive rate
//	bf := bloom.NewWithEstimates(n, fp)
//
//	// Add the public keys to the filter
//	if pubKey1 != "" {
//		bf.Add([]byte(pubKey1))
//	}
//	if pubKey2 != "" {
//		bf.Add([]byte(pubKey2))
//	}
//	if pubKey3 != "" {
//		bf.Add([]byte(pubKey3))
//	}
//
//	bf.K()
//
//	// print the value fo k
//	fmt.Println(bloom.EstimateParameters(n, fp))
//
//	// Test for existence (false positive)
//	if bf.Test([]byte(pubKey1)) {
//		fmt.Println("pubKey1 Exists!")
//	}
//
//	if bf.Test([]byte(pubKey2)) {
//		fmt.Println("pubKey2 Exists!")
//	}
//
//	if bf.Test([]byte(pubKey3)) {
//		fmt.Println("pubKey3 Exists!")
//	}
//
//	//// Serialize the Bloom filter
//	//serializedBloomFilter, err := bf.MarshalJSON()
//	//if err != nil {
//	//	log.Fatalf("Failed to serialize Bloom filter: %v", err)
//	//}
//
//	// Serialize the Bloom filter
//	// Assuming `b` is your bloom filter
//	var buf bytes.Buffer
//	_, err := bf.WriteTo(&buf)
//	if err != nil {
//		log.Fatalf("Error serializing the Bloom filter: %v", err)
//	}
//
//	serializedBloomFilter := buf.Bytes()
//
//	return serializedBloomFilter
//}
