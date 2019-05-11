package data

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

	bc "../../blockchain"
	"../../mpt"
)

var ContractAddress = "636f6e7472616374"

//Transaction will be handle and validate by Miner
type Transaction struct {
	ID          string `json:"id"`
	FromAddress string `json:"fromAddress"` //PublicKey
	ToAddress   string `json:"toAddress"`   //PublicKey
	Value       int    `json:"value"`       //Money
	Data        string `json:"data"`        //Data of the game
}

//SignedTransaction will be send between nodes in Network
type SignedTransaction struct {
	Transaction Transaction `json:"transaction"`
	Signature   []byte      `json:"signature"`
}

//DecodeTransactionFromJSON function takes a string represents the json value of a Transaction,
//decode the string back to a Transaction instance
func DecodeTransactionFromJSON(jsonString string) (Transaction, error) {
	var result Transaction

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

//EncodeToJSON function encode a Transaction instance into json string
func (transaction *Transaction) EncodeToJSON() (string, error) {
	var result string

	transactionByte, err := json.Marshal(&transaction)
	if err != nil {
		return result, err
	}
	result = string(transactionByte)
	return result, nil
}

//DecodeSignedTransactionFromJSON function takes a string represents the json value of a SignedTransaction,
//decode the string back to a SignedTransaction instance
func DecodeSignedTransactionFromJSON(jsonString string) (SignedTransaction, error) {
	var result SignedTransaction

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

//EncodeToJSON function encode a Transaction instance into json string
func (signedTransaction *SignedTransaction) EncodeToJSON() (string, error) {
	var result string

	signedTransactionByte, err := json.Marshal(&signedTransaction)
	if err != nil {
		return result, err
	}
	result = string(signedTransactionByte)
	return result, nil
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	hash = sha256.Sum256(tx.Serialize())

	return hash[:]
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}
	return transaction
}

func GetSignedTxsFromMPT(txsTrie mpt.MerklePatriciaTrie) []SignedTransaction {
	var result []SignedTransaction
	for key := range txsTrie.KeyVal {
		signedTxJSON, err := txsTrie.Get(key)
		if err != nil {
			panic(err)
		} else {
			signedTx, _ := DecodeSignedTransactionFromJSON(signedTxJSON)
			result = append(result, signedTx)
		}
	}
	return result
}

func AddTransaction(accountTrie mpt.MerklePatriciaTrie,
	signedTxsTrie mpt.MerklePatriciaTrie) mpt.MerklePatriciaTrie {
	signedTxs := GetSignedTxsFromMPT(signedTxsTrie)
	for _, signedTx := range signedTxs {
		tx := signedTx.Transaction
		fmt.Println(string([]rune(tx.ToAddress)[:len(ContractAddress)]))
		if string([]rune(tx.ToAddress)[:len(ContractAddress)]) == ContractAddress {
			fmt.Println("Handle game contract")
			gameAccountJSON, _ := accountTrie.Get(ContractAddress)
			gameAccount, _ := bc.DecodeAccountFromJSON(gameAccountJSON)
			gameList, _ := DecodeGameListFromJSON(gameAccount.Data)
			gameList, ok := HandleGameContract(tx, gameList)
			if ok {
				gameListJSON, _ := EncodeGameListToJSON(gameList)
				gameAccount.Data = gameListJSON
				gameAccountJSON, _ := gameAccount.EncodeToJSON()
				//Change function address to contract address to update balance
				tx.ToAddress = ContractAddress
				accountTrie.Insert(tx.ToAddress, gameAccountJSON)
				Rebalance(&accountTrie, tx)
			}
		} else { //Transfer money contract
			Rebalance(&accountTrie, tx)
		}

	}
	return accountTrie
}

func Rebalance(accountTrie *mpt.MerklePatriciaTrie, tx Transaction) {
	FromAccountJSON, _ := accountTrie.Get(tx.FromAddress)
	ToAccountJSON, _ := accountTrie.Get(tx.ToAddress)
	FromAccount, _ := bc.DecodeAccountFromJSON(FromAccountJSON)
	ToAccount, _ := bc.DecodeAccountFromJSON(ToAccountJSON)
	FromAccount.Balance = FromAccount.Balance - tx.Value
	ToAccount.Balance = ToAccount.Balance + tx.Value
	FromAccountJSON, _ = FromAccount.EncodeToJSON()
	ToAccountJSON, _ = ToAccount.EncodeToJSON()
	accountTrie.Insert(tx.FromAddress, FromAccountJSON)
	accountTrie.Insert(tx.ToAddress, ToAccountJSON)
}
