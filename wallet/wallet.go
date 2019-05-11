package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"log"

	"../network/data"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

func Address(pubKey []byte) []byte {
	pubHash := PublicKeyHash(pubKey)

	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	return address
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
func ValidateTxSignature(transaction data.Transaction, signature []byte) (bool, error) {
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

func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}

	publicRipMD := hasher.Sum(nil)

	return publicRipMD
}

func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}

func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLength]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)

	return []byte(encode)
}

func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input[:]))
	if err != nil {
		log.Panic(err)
	}

	return decode
}
