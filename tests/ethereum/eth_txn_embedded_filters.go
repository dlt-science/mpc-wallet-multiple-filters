package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	cuckoofilter "github.com/panmari/cuckoofilter"
	"log"
	"math/big"
	"os"
	"strconv"
)

func main() {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	//// Load the public keys from environment variables
	//pubKey1 := os.Getenv("PUB_KEY_1")
	//pubKey2 := os.Getenv("PUB_KEY_2")
	//pubKey3 := os.Getenv("PUB_KEY_3")

	//// Load the private keys from environment variables
	//Pvk1 := os.Getenv("PVK_1")
	//Pvk2 := os.Getenv("PVK_2")
	//Pvk3 := os.Getenv("PVK_3")
	//
	// Load the remaning env variables
	PrivateKey := os.Getenv("PRIVATE_KEY")
	//FROM_ADDRESS := os.Getenv("FROM_ADDRESS")
	ToAddress := os.Getenv("TO_ADDRESS")
	//
	//pubKey1, pubKey2, pubKey3 := genPubKeys(Pvk1, Pvk2, Pvk3)

	// Define the party sets to test
	partySets := [][]int{
		{1, 1},
		{2, 2},
		{2, 3},
		{3, 4},
		{4, 5},
		{5, 10},
	}

	// Connect to the Ethereum client
	EthNodeUrl := os.Getenv("ETH_NODE_URL")
	if EthNodeUrl == "" {
		// Use Infura's Ethereum Goerly testnet public endpoint as the default
		// https://www.alchemy.com/chain-connect/chain/goerli
		EthNodeUrl = "https://rpc2.sepolia.org/"
	}

	client, err := ethclient.Dial(EthNodeUrl)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// Load the private key from environment variable
	privateKey, err := crypto.HexToECDSA(PrivateKey)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// Check the accounts balance
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	balance, err := client.BalanceAt(context.Background(), fromAddress, nil)
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}
	fmt.Printf("Balance: %s\n", balance.String())

	// Create the transaction
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to suggest gas price: %v", err)
	}

	//// Set the gas limit based on the SuggestGasPrice * (1 + gas buffer)
	//gasLimit := gasPrice.Mul(gasPrice, big.NewInt(1.2))
	gasLimit := getGasLimit(gasPrice)
	println("gasLimit to send transaction with payload: ", gasLimit)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get network ID: %v", err)
	}

	// Define list of csv documents to be created:
	type FilterFunc func(publicKeys []string, n uint, fp float64) []byte

	type FuncFilePair struct {
		Func FilterFunc
		File string
	}

	funcFilePairs := []FuncFilePair{
		{getBloomFilter, "EthereumBloomResults.csv"},
		{getCuckooFilter, "EthereumCuckooResults.csv"},
	}

	for _, pair := range funcFilePairs {
		fileName := pair.File
		filterFunc := pair.Func

		// Create a new CSV file
		file, err := os.Create(fileName)
		if err != nil {
			log.Fatalf("Failed to create CSV file: %v", err)
		}
		defer file.Close()

		// Create a new CSV writer
		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write the header row to the CSV file
		err = writer.Write([]string{"partySet", "filterSize", "transactionSize", "transactionFeeETH", "transactionFeeWei"})
		if err != nil {
			log.Fatalf("Failed to write to CSV file: %v", err)
		}

		for _, partySet := range partySets {

			// Generate a set of private keys
			privateKeys := generatePrivateKeys(partySet[0])

			// Generate a set of public keys
			publicKeys := generatePublicKeys(privateKeys)

			// Create a new Bloom Filter with 1,000,000 items and 0.001% false positive rate
			//serializedBloomFilter := getBloomFilter(pubKey1, pubKey2, pubKey3, 3, 0.0001)
			//serializedBloomFilter := getBloomFilter(pubKey1, pubKey2, pubKey3, 3, 0.03)
			//serializedBloomFilter := getBloomFilter(publicKeys, uint(partySet[1]), 0.0001)
			serializedBloomFilter := filterFunc(publicKeys, uint(partySet[1]), 0.0001)

			// Recipient address and amount for test transaction
			toAddress := common.HexToAddress(ToAddress)

			//// Sent 1 ETH to the recipient address
			//amount := big.NewInt(1000000000000000000) // 1 ETH

			//// Sent 0.01 ETH to the recipient address
			//amount := big.NewInt(10000000000000000) // 0.01 ETH

			// Sent 0.0001 ETH to the recipient address
			amount := big.NewInt(100000000000000) // 0.0001 ETH

			// Get the nonce for the account sending the transaction
			nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
			if err != nil {
				log.Fatalf("Failed to get nonce: %v", err)
			}

			// Convert the gas limit to an integer 64

			txData := &types.LegacyTx{
				Nonce:    nonce,
				GasPrice: gasPrice,
				Gas:      gasLimit,
				To:       &toAddress,
				Value:    amount,
				Data:     serializedBloomFilter,
			}

			tx := types.NewTx(txData)

			//tx := types.NewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, serializedBloomFilter)

			// Sign the transaction
			signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
			if err != nil {
				log.Fatalf("Failed to sign transaction: %v", err)
			}

			txSize := getTxSize(err, signedTx)
			fmt.Println("The transaction size in bytes is: ", txSize)

			// Get the gas needed to send the transaction
			gasNeededInWei := gasNeeded(fromAddress, toAddress, amount, serializedBloomFilter, err, client)
			println("gasNeeded to send the transaction of size ", txSize, " is: ", gasNeededInWei, " wei")

			// Calculate the transaction fee in ETH
			txFeeInETH := float64(gasNeededInWei) / 1000000000000000000
			println("The transaction fee is: ", txFeeInETH, " ETH")

			// Write the results to the CSV file
			err = writer.Write([]string{
				fmt.Sprintf("%d of %d", partySet[0], partySet[1]),
				strconv.Itoa(len(serializedBloomFilter)),
				strconv.Itoa(txSize),
				fmt.Sprintf("%.16f", txFeeInETH),
				strconv.Itoa(int(gasNeededInWei)),
			})

			if err != nil {
				log.Fatalf("Failed to write to CSV file: %v", err)
			}

		}

	}

}

func getGasLimit(gasPrice *big.Int) uint64 {
	/**
	 * The gas limit is the maximum amount of gas that can be used in a transaction.
	 * The gas limit is set to 21000 for a simple send transaction.
	 * The gas limit is set to 53000 for a simple send transaction with a data payload.
	 */

	// Multiply the gas price by 1.2
	gasLimitBigFloat := new(big.Float).Mul(new(big.Float).SetInt(gasPrice), big.NewFloat(1.2))

	// Convert the big.Float to a big.Int
	gasLimitBigInt := new(big.Int)
	gasLimitBigFloat.Int(gasLimitBigInt)

	// Convert the big.Int to a uint64
	gasLimit := gasLimitBigInt.Uint64()
	return gasLimit
}

func gasNeeded(fromAddress common.Address, toAddress common.Address, amount *big.Int, serializedBloomFilter []byte, err error, client *ethclient.Client) uint64 {
	// Create a CallMsg
	callMsg := ethereum.CallMsg{
		From:     fromAddress,
		To:       &toAddress,
		Gas:      0,   // Set to 0 for automatic estimation
		GasPrice: nil, // Not used for gas estimation
		Value:    amount,
		Data:     serializedBloomFilter,
	}

	// Estimate the gas needed for the transaction
	gasLimit, err := client.EstimateGas(context.Background(), callMsg)
	if err != nil {
		log.Fatalf("Failed to estimate gas: %v", err)
	}

	fmt.Printf("Estimated gas needed for signed transaction with Payload: %d\n", gasLimit)
	return gasLimit
}

func getTxSize(err error, signedTx *types.Transaction) int {
	// Serialize the transaction
	data, err := signedTx.MarshalBinary()
	if err != nil {
		log.Fatalf("Failed to serialize transaction: %v", err)
	}

	// Measure the transaction size
	txSize := len(data)
	fmt.Printf("Transaction size: %d bytes\n", txSize)

	return txSize
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

func getCuckooFilter(publicKeys []string, n uint, fp float64) []byte {

	cf := cuckoofilter.NewFilter(n)

	for _, pubKey := range publicKeys {
		//_, pubKey := btcec.PrivKeyFromBytes([]byte(privateKey))
		//pubKeyBytes := pubKey.SerializeCompressed()
		pubKeyBytes, _ := hex.DecodeString(pubKey)
		cf.Insert(pubKeyBytes)

		if cf.Lookup(pubKeyBytes) {
			fmt.Println(pubKeyBytes, " Exists!")
		}
	}

	serializedCuckooFilter := cf.Encode()

	return serializedCuckooFilter
}

//func genPubKeys(PVK_1 string, PVK_2 string, PVK_3 string) (bytes string, bytes2 string, bytes3 string) {
//	// Generate the public keys from the private keys
//	privKey1, err := crypto.HexToECDSA(PVK_1)
//	if err != nil {
//		log.Fatalf("Failed to load private key: %v", err)
//	}
//	pubKey1 := privKey1.PublicKey
//	//pubKey11 := privKey1.Public()
//	//
//	//fmt.Println("pubKey11: ", pubKey11)
//
//	privKey2, err := crypto.HexToECDSA(PVK_2)
//	if err != nil {
//		log.Fatalf("Failed to load private key: %v", err)
//	}
//	pubKey2 := privKey2.PublicKey
//
//	privKey3, err := crypto.HexToECDSA(PVK_3)
//	if err != nil {
//		log.Fatalf("Failed to load private key: %v", err)
//	}
//	pubKey3 := privKey3.PublicKey
//
//	//// Convert the public keys to strings
//	//pubKey1String := publicKeyToString(pubKey1)
//	//pubKey2String := publicKeyToString(pubKey2)
//	//pubKey3String := publicKeyToString(pubKey3)
//
//	//return pubKey1String, pubKey2String, pubKey3String
//
//	// Convert the public keys to bytes
//	pubKey1Bytes := crypto.FromECDSAPub(&pubKey1)
//	pubKey2Bytes := crypto.FromECDSAPub(&pubKey2)
//	pubKey3Bytes := crypto.FromECDSAPub(&pubKey3)
//
//	return hex.EncodeToString(pubKey1Bytes), hex.EncodeToString(pubKey2Bytes), hex.EncodeToString(pubKey3Bytes)
//}

//func getBloomFilter(pubKey1 string, pubKey2 string, pubKey3 string, n uint, fp float64) []byte {
//
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
