package models

import "encoding/json"

//Account use to keep username and password
type Account struct {
	Username  string
	Password  string
	PublicKey []byte
}

//CheckAccount will validate username and password
func (account *Account) CheckAccount(username string, password string) bool {
	return username == account.Username && password == account.Password
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
