package wallet

import (
	"crypto/ecdsa"
	"encoding/hex"
	"log"

	"../blockchain"

	"github.com/ethereum/go-ethereum/crypto"
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

//SignData will create a signed data base on private key
func SignData(dataString string, key string) ([]byte, error) {
	//Generate public key
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return nil, err
	}
	data := []byte(dataString)
	hash := crypto.Keccak256Hash(data)

	//Sign data
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return signature, nil
}

func ValidateSignature(data string, signature []byte, publicKey string) (bool, error) {
	dataByte := []byte(data)
	hash := crypto.Keccak256Hash(dataByte)
	sigPublicKey, err := crypto.Ecrecover(hash.Bytes(), signature)
	if err != nil {
		return false, err
	}
	matches := hex.EncodeToString(sigPublicKey) == publicKey

	return matches, nil
}

//ValidateSignature will check if message was not changed
func ValidateTransaction(transaction blockchain.Transaction, signature []byte) (bool, error) {
	transactionJSON, err := transaction.EncodeToJSON()
	if err != nil {
		return false, err
	}
	data := []byte(transactionJSON)
	hash := crypto.Keccak256Hash(data)

	sigPublicKey, err := crypto.Ecrecover(hash.Bytes(), signature)
	if err != nil {
		return false, err
	}

	matches := hex.EncodeToString(sigPublicKey) == transaction.FromAddress

	return matches, nil
}

//GenerateKey will return a private, public key
func GenerateKey() ([]byte, []byte) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)

	publicKeyBytes := GetPublicKey(privateKey)
	return privateKeyBytes, publicKeyBytes
}

func GetPublicKey(privateKey *ecdsa.PrivateKey) []byte {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	return publicKeyBytes
}
