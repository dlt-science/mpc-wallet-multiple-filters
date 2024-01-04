package main

import (
	"bytes"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"strconv"

	//"github.com/btcsuite/btcd/btcutil"
	//"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/joho/godotenv"
	cuckoofilter "github.com/panmari/cuckoofilter"
	"log"
	"os"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Create a new CSV file
	file, err := os.Create("BitcoinCuckooResults.csv")
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

	for _, partySet := range partySets {
		// Generate a set of private keys
		privateKeys := generatePrivateKeys(partySet[0])

		// Generate a set of public keys
		publicKeys := generatePublicKeys(privateKeys)

		// Generate a Cuckoo filter for the private keys
		// Generating a filter for total accepted items for a set. E.g. filter of 4 which would have 3 items
		serializedCuckooFilter := getCuckooFilter(publicKeys, partySet[1], 0.0001)

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
		txOutValue := msgTx.TxOut[0].Value
		println("The transaction output value is: ", txOutValue, " satoshis")

		// Calculate the transaction fee
		txOutValueInBTC := float64(txOutValue) / 100000000
		println("The transaction fee is: ", txOutValueInBTC, " BTC")

		// Write the results to the CSV file
		err = writer.Write([]string{
			fmt.Sprintf("%d of %d", partySet[0], partySet[1]),
			strconv.Itoa(len(serializedCuckooFilterBytes)),
			strconv.Itoa(txSize),
			fmt.Sprintf("%.8f", txOutValueInBTC),
			strconv.Itoa(int(txOutValue)),
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

func getCuckooFilter(privateKeys []string, m int, fp float64) []byte {
	cf := cuckoofilter.NewFilter(uint(len(privateKeys)))

	for _, privateKey := range privateKeys {
		_, pubKey := btcec.PrivKeyFromBytes([]byte(privateKey))
		pubKeyBytes := pubKey.SerializeCompressed()
		cf.Insert(pubKeyBytes)

		if cf.Lookup([]byte(privateKey)) {
			fmt.Println(privateKey, " Exists!")
		}
	}

	serializedCuckooFilter := cf.Encode()

	return serializedCuckooFilter
}

//func getCuckooFilter(pubKey1 string, pubKey2 string, pubKey3 string, n uint, fp float64) []byte {
//	//// Create a new Cuckoo Filter with n items, fp false positive rate and
//	////a default 4 number of entries or fingerprints per bucket
//	////based on the paper https://www.pdl.cmu.edu/PDL-FTP/FS/cuckoo-conext2014.pdf
//	//cf := cuckoo.NewCuckooFilter(n, fp, 4)
//
//	cf := cuckoofilter.NewFilter(n)
//
//	// Add the public keys to the filter
//	if pubKey1 != "" {
//		cf.Insert([]byte(pubKey1))
//	}
//	if pubKey2 != "" {
//		cf.Insert([]byte(pubKey2))
//	}
//	if pubKey3 != "" {
//		cf.Insert([]byte(pubKey3))
//	}
//
//	// Check if the public keys are present in the filter
//	if cf.Lookup([]byte(pubKey1)) {
//		fmt.Println("pubKey1 Exists!")
//	}
//	if cf.Lookup([]byte(pubKey2)) {
//		fmt.Println("pubKey2 Exists!")
//	}
//	if cf.Lookup([]byte(pubKey3)) {
//		fmt.Println("pubKey3 Exists!")
//	}
//
//	//// Serialize the Cuckoo filter
//	//serializedCuckooFilter, err := cf.Encode()
//	//if err != nil {
//	//	log.Fatalf("Failed to serialize Cuckoo filter: %v", err)
//	//}
//
//	// Serialize the Cuckoo filter
//	serializedCuckooFilter := cf.Encode()
//
//	return serializedCuckooFilter
//}
