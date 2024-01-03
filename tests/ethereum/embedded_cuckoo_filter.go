package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"

	//cuckoofilter "github.com/seiflotfy/cuckoofilter" // for 0.03% false positive rate
	cuckoofilter "github.com/panmari/cuckoofilter" // for 0.01% false positive rate or 0.0001
	//cuckoofilter "github.com/irfansharif/cfilter" // flexibility to set false positive rate
	//"cuckoo" // flexibility to set false positive rate

	//"github.com/seiflotfy/cuckoofilter"
	"log"
	"math/big"
	"os"
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

	// Load the private keys from environment variables
	PVK_1 := os.Getenv("PVK_1")
	PVK_2 := os.Getenv("PVK_2")
	PVK_3 := os.Getenv("PVK_3")

	// Load the remaning env variables
	PRIVATE_KEY := os.Getenv("PRIVATE_KEY")
	//FROM_ADDRESS := os.Getenv("FROM_ADDRESS")
	TO_ADDRESS := os.Getenv("TO_ADDRESS")

	pubKey1, pubKey2, pubKey3 := genPubKeys(PVK_1, PVK_2, PVK_3)

	serializedCuckooFilter := getCuckooFilter(pubKey1, pubKey2, pubKey3, 3, 0.0001)

	// Connect to the Ethereum client
	ETH_NODE_URL := os.Getenv("ETH_NODE_URL")
	if ETH_NODE_URL == "" {
		// Use Infura's Ethereum Goerly testnet public endpoint as the default
		// https://www.alchemy.com/chain-connect/chain/goerli
		ETH_NODE_URL = "https://rpc2.sepolia.org/"
	}

	client, err := ethclient.Dial(ETH_NODE_URL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// Load the private key from environment variable
	privateKey, err := crypto.HexToECDSA(PRIVATE_KEY)
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

	// Recipient address and amount for test transaction
	toAddress := common.HexToAddress(TO_ADDRESS)

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

	// Create the transaction
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to suggest gas price: %v", err)
	}

	//// Set the gas limit based on the SuggestGasPrice * (1 + gas buffer)
	//gasLimit := gasPrice.Mul(gasPrice, big.NewInt(1.2))
	gasLimit := getGasLimit(gasPrice)
	println("gasLimit to send transaction with payload: ", gasLimit)

	// Convert the gas limit to an integer 64

	txData := &types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddress,
		Value:    amount,
		Data:     serializedCuckooFilter,
	}

	tx := types.NewTx(txData)

	//tx := types.NewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, serializedBloomFilter)

	// Sign the transaction
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get network ID: %v", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	txSize := getTxSize(err, signedTx)
	fmt.Println("The transaction size in bytes is: ", txSize)

	// Get the gas needed to send the transaction
	gasNeeded := gasNeeded(fromAddress, toAddress, amount, serializedCuckooFilter, err, client)
	println("gasNeeded to send the transaction of size ", txSize, " is: ", gasNeeded)

	//// Record the time before sending the transaction
	//startTime := time.Now()
	//
	//// Broadcast the transaction
	//err = client.SendTransaction(context.Background(), signedTx)
	//if err != nil {
	//	log.Fatalf("Failed to send transaction: %v", err)
	//}
	//
	//// Record the time after the transaction is sent
	//endTime := time.Now()
	//
	//fmt.Printf("Transaction sent! TX Hash: %s\n", signedTx.Hash().Hex())
	//
	//// Calculate the validation time
	//validationTime := endTime.Sub(startTime)
	//
	//fmt.Println("The validation time is: ", validationTime)

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

func genPubKeys(PVK_1 string, PVK_2 string, PVK_3 string) (bytes string, bytes2 string, bytes3 string) {
	// Generate the public keys from the private keys
	privKey1, err := crypto.HexToECDSA(PVK_1)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}
	pubKey1 := privKey1.PublicKey
	//pubKey11 := privKey1.Public()
	//
	//fmt.Println("pubKey11: ", pubKey11)

	privKey2, err := crypto.HexToECDSA(PVK_2)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}
	pubKey2 := privKey2.PublicKey

	privKey3, err := crypto.HexToECDSA(PVK_3)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}
	pubKey3 := privKey3.PublicKey

	//// Convert the public keys to strings
	//pubKey1String := publicKeyToString(pubKey1)
	//pubKey2String := publicKeyToString(pubKey2)
	//pubKey3String := publicKeyToString(pubKey3)

	//return pubKey1String, pubKey2String, pubKey3String

	// Convert the public keys to bytes
	pubKey1Bytes := crypto.FromECDSAPub(&pubKey1)
	pubKey2Bytes := crypto.FromECDSAPub(&pubKey2)
	pubKey3Bytes := crypto.FromECDSAPub(&pubKey3)

	return hex.EncodeToString(pubKey1Bytes), hex.EncodeToString(pubKey2Bytes), hex.EncodeToString(pubKey3Bytes)
}

func getCuckooFilter(pubKey1 string, pubKey2 string, pubKey3 string, n uint, fp float64) []byte {
	//// Create a new Cuckoo Filter with n items, fp false positive rate and
	////a default 4 number of entries or fingerprints per bucket
	////based on the paper https://www.pdl.cmu.edu/PDL-FTP/FS/cuckoo-conext2014.pdf
	//cf := cuckoo.NewCuckooFilter(n, fp, 4)

	cf := cuckoofilter.NewFilter(n)

	// Add the public keys to the filter
	if pubKey1 != "" {
		cf.Insert([]byte(pubKey1))
	}
	if pubKey2 != "" {
		cf.Insert([]byte(pubKey2))
	}
	if pubKey3 != "" {
		cf.Insert([]byte(pubKey3))
	}

	// Check if the public keys are present in the filter
	if cf.Lookup([]byte(pubKey1)) {
		fmt.Println("pubKey1 Exists!")
	}
	if cf.Lookup([]byte(pubKey2)) {
		fmt.Println("pubKey2 Exists!")
	}
	if cf.Lookup([]byte(pubKey3)) {
		fmt.Println("pubKey3 Exists!")
	}

	//// Serialize the Cuckoo filter
	//serializedCuckooFilter, err := cf.Encode()
	//if err != nil {
	//	log.Fatalf("Failed to serialize Cuckoo filter: %v", err)
	//}

	// Serialize the Cuckoo filter
	serializedCuckooFilter := cf.Encode()

	return serializedCuckooFilter
}
