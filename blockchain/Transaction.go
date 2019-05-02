package blockchain

import "encoding/json"

//Transaction will be handle and validate by Miner
type Transaction struct {
	ID          string
	FromAddress string  `json:"from"`  //PublicKey
	ToAddress   string  `json:"to"`    //PublicKey
	Value       float32 `json:"value"` //Money
	Data        string  `json:"data"`  //Data of the game
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
