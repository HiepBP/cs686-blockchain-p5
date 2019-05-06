package models

import "encoding/json"

type UserMsg struct {
	Signature string `json:"signature"`
	Data      string `json:"data"`
}

func DecodeUserMsgFromJSON(jsonString string) (UserMsg, error) {
	var result UserMsg

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (data *UserMsg) EncodeToJSON() (string, error) {
	var result string

	dataByte, err := json.Marshal(&data)
	if err != nil {
		return result, err
	}
	result = string(dataByte)
	return result, nil
}
