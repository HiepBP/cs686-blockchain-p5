package models

import "encoding/json"

type Data struct {
	FunctionName string
	Args         string
}

func DecodeDataFromJSON(jsonString string) (Data, error) {
	var result Data

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (data *Data) EncodeToJSON() (string, error) {
	var result string

	dataByte, err := json.Marshal(&data)
	if err != nil {
		return result, err
	}
	result = string(dataByte)
	return result, nil
}
