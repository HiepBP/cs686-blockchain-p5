package data

import "encoding/json"

//RegisterData is
type RegisterData struct {
	AssignedId  int32  `json:"assignedId"`
	PeerMapJson string `json:"peerMapJson"`
}

//NewRegisterData is a constructor
func NewRegisterData(id int32, peerMapJSON string) RegisterData {
	return RegisterData{
		AssignedId:  id,
		PeerMapJson: peerMapJSON,
	}
}

//EncodeToJson return json string for RegisterData
func (data *RegisterData) EncodeToJson() (string, error) {
	var result string
	registerDataByte, err := json.Marshal(data)
	if err != nil {
		return result, err
	}
	result = string(registerDataByte)
	return result, nil
}
