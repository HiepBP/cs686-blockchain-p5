package bc

import "encoding/json"

type Account struct {
	Balance int    `json:"balance"`
	Data    string `json:"data"`
}

func DecodeAccountFromJSON(jsonString string) (Account, error) {
	var result Account

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (account *Account) EncodeToJSON() (string, error) {
	var result string

	accountByte, err := json.Marshal(&account)
	if err != nil {
		return result, err
	}
	result = string(accountByte)
	return result, nil
}
