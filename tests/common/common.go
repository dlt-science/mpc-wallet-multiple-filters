package common

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

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
