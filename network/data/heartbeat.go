package data

import (
	"encoding/json"

	"../../blockchain"
)

type HeartBeatData struct {
	IfNewBlock  bool   `json:"ifNewBlock"`
	Id          int32  `json:"id"`
	BlockJson   string `json:"blockJson"`
	PeerMapJson string `json:"peerMapJson"`
	Addr        string `json:"addr"`
	Hops        int32  `json:"hops"`
}

//NewHeartBeatData is a constructor
func NewHeartBeatData(ifNewBlock bool, id int32, blockJSON string, peerMapJSON string, addr string) HeartBeatData {
	return HeartBeatData{
		IfNewBlock:  ifNewBlock,
		Id:          id,
		BlockJson:   blockJSON,
		PeerMapJson: peerMapJSON,
		Addr:        addr,
		Hops:        2,
	}
}

//PrepareHeartBeatData will create heart beat base on sbc is empty or not
func PrepareHeartBeatData(sbc *SyncBlockChain, selfID int32, block bc.Block, peerMapJSON string, addr string) HeartBeatData {
	if block.Header.Hash != "" {
		sbc.Insert(block)
		blockJSON, _ := block.EncodeToJSON()
		return NewHeartBeatData(true, selfID, blockJSON, peerMapJSON, addr)
	}
	return NewHeartBeatData(false, selfID, "", peerMapJSON, addr)
}

//DecodeHeartBeatFromJSON will decode to HeartBeatData from jsonString
func DecodeHeartBeatFromJSON(jsonString string) (HeartBeatData, error) {

	var result HeartBeatData

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

//EncodeToJSON function encode a block instance into json string
func (hb *HeartBeatData) EncodeToJSON() (string, error) {
	var result string

	hbByte, err := json.Marshal(&hb)
	if err != nil {
		return result, err
	}
	result = string(hbByte)
	return result, nil
}
