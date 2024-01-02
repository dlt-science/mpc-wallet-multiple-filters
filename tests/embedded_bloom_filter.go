package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
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

	serializedBloomFilter := getFilter(pubKey1, pubKey2, pubKey3, 3, 0.01)

	// Connect to the Ethereum client
	ETH_NODE_URL := os.Getenv("ETH_NODE_URL")
	if ETH_NODE_URL == "" {
		// Use Infura's Ethereum Goerly testnet public endpoint as the default
		// https://www.alchemy.com/chain-connect/chain/goerli
		ETH_NODE_URL = "https://goerli.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161"
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

	// Recipient address and amount for test transaction
	toAddress := common.HexToAddress(TO_ADDRESS)

	//// Sent 1 ETH to the recipient address
	//amount := big.NewInt(1000000000000000000) // 1 ETH

	// Sent 0.01 ETH to the recipient address
	amount := big.NewInt(10000000000000000) // 0.01 ETH

	// Get the nonce for the account sending the transaction
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}

	// Create the transaction
	gasLimit := uint64(21000) // Gas limit for standard transaction
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to suggest gas price: %v", err)
	}

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
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get network ID: %v", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	// Broadcast the transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}

	fmt.Printf("Transaction sent! TX Hash: %s\n", signedTx.Hash().Hex())

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

//func genPrivatePublicKeys() {
// Generate a new random key pair
// priv, pub, err := box.GenerateKey(rand.Reader)
// if err != nil {
// 	log.Fatal(err)
// }
// fmt.Printf("Private Key: %x\n", priv[:])
// fmt.Printf("Public Key: %x\n", pub[:])
//}

func getFilter(pubKey1 string, pubKey2 string, pubKey3 string, n uint, fp float64) []byte {

	// Check if n and fp are zero, and if so, assign them default values
	if n == 0 {
		n = 3 // default value
	}
	if fp == 0 {
		fp = 0.01 // default value
	}

	// Create a new Bloom Filter with 1,000,000 items and 0.01% false positive rate
	bf := bloom.NewWithEstimates(3, 0.01)

	// Add the public keys to the filter
	if pubKey1 != "" {
		bf.Add([]byte(pubKey1))
	}
	if pubKey2 != "" {
		bf.Add([]byte(pubKey2))
	}
	if pubKey3 != "" {
		bf.Add([]byte(pubKey3))
	}

	bf.K()

	// print the value fo k
	fmt.Println(bloom.EstimateParameters(3, 0.01))

	// Test for existence (false positive)
	if bf.Test([]byte(pubKey1)) {
		fmt.Println("Exists!")
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
