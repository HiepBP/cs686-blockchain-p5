package models

import "encoding/json"

type Game struct {
	ID           uint32
	Dealer       string
	DealerChoice uint32
	DealerHash   string
	Player       string
	PlayerChoice uint32
	GameValue    float32
	Result       uint32
	Closed       bool
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
