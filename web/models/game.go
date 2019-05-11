package models

import "encoding/json"

type Game struct {
	ID           uint32 `json:"id"`
	Dealer       string `json:"dealer"`
	DealerChoice uint32 `json:"dealerChoice"`
	DealerHash   string `json:"dealerHash"`
	Player       string `json:"plaer"`
	PlayerChoice uint32 `json:"playerChoice"`
	GameValue    int    `json:"gameValue"`
	Result       uint32 `json:"result"`
	Closed       bool   `json:"closed"`
}

func DecodeGameFromJSON(jsonString string) (Game, error) {
	var result Game

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (game *Game) EncodeToJSON() (string, error) {
	var result string

	gameByte, err := json.Marshal(&game)
	if err != nil {
		return result, err
	}
	result = string(gameByte)
	return result, nil
}
