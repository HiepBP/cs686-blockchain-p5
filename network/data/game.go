package data

import "encoding/json"

type Game struct {
	ID           uint32 `json:"id"`
	Dealer       string `json:"dealer"`
	DealerChoice uint32 `json:"dealerChoice"`
	DealerHash   string `json:"dealerHash"`
	DealerValue    int    `json:"dealerValue"`
	Player       string `json:"player"`
	PlayerChoice uint32 `json:"playerChoice"`
	PlayerValue int `json:"playerValue"`
	Result       uint32 `json:"result"`
	Closed       bool   `json:"closed"`
}

type GameCreate struct {
	DealerHash string `json:"dealerHash"`
}

type GameJoin struct {
	ID           uint32 `json:"id"`
	PlayerChoice uint32 `json:"playerChoice"`
}

type GameReveal struct {
	ID           uint32 `json:"ID"`
	DealerChoice uint32 `json:"DealerChoice"`
	SecretNumber string `json:"SecretNumber"`
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

func DecodeGameCreateFromJSON(jsonString string) (GameCreate, error) {
	var result GameCreate

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (gameCreate *GameCreate) EncodeToJSON() (string, error) {
	var result string

	gameCreateByte, err := json.Marshal(&gameCreate)
	if err != nil {
		return result, err
	}
	result = string(gameCreateByte)
	return result, nil
}

func DecodeGameJoinFromJSON(jsonString string) (GameJoin, error) {
	var result GameJoin

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (gameJoin *GameJoin) EncodeToJSON() (string, error) {
	var result string

	gameJoinByte, err := json.Marshal(&gameJoin)
	if err != nil {
		return result, err
	}
	result = string(gameJoinByte)
	return result, nil
}

func DecodeGameRevealFromJSON(jsonString string) (GameReveal, error) {
	var result GameReveal

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (gameReveal *GameReveal) EncodeToJSON() (string, error) {
	var result string

	gameRevealByte, err := json.Marshal(&gameReveal)
	if err != nil {
		return result, err
	}
	result = string(gameRevealByte)
	return result, nil
}
