package data

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"golang.org/x/crypto/sha3"
)

var (
	CreateGameAddress   = "636f6e7472616374637265617465"
	JoinGameAddress     = "636f6e74726163746a6f696e"
	RevealChoiceAddress = "636f6e747261637472657665616c"
)

func DecodeGameListFromJSON(jsonString string) ([]Game, error) {

	var result []Game

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func EncodeGameListToJSON(gameList []Game) (string, error) {
	var result string

	gamesByte, err := json.Marshal(&gameList)
	if err != nil {
		return result, err
	}
	result = string(gamesByte)
	return result, nil
}

func CreateGame(tx Transaction, gameList []Game) []Game {
	gameCreate, _ := DecodeGameCreateFromJSON(tx.Data)
	var game Game
	game.ID = uint32(len(gameList))
	game.Dealer = tx.FromAddress
	game.DealerHash = gameCreate.DealerHash
	game.DealerChoice = 0
	game.GameValue = tx.Value
	gameList = append(gameList, game)

	return gameList
}

func JoinGame(tx Transaction, gameList []Game) []Game {
	gameJoin, _ := DecodeGameJoinFromJSON(tx.Data)
	game := gameList[gameJoin.ID]
	game.Player = tx.FromAddress
	game.PlayerChoice = gameJoin.PlayerChoice
	fmt.Println(gameList[gameJoin.ID].PlayerChoice)
	return gameList
}

func RevealGame(tx Transaction, gameList []Game) ([]Game, bool) {
	gameReveal, _ := DecodeGameRevealFromJSON(tx.Data)
	game := gameList[gameReveal.ID]
	dealerHash := sha3.Sum256([]byte(strconv.Itoa(int(gameReveal.DealerChoice)) +
		strconv.Itoa(int(gameReveal.SecretNumber))))
	if hex.EncodeToString(dealerHash[:]) != game.DealerHash {
		fmt.Println("Incorrect key")
		return gameList, false
	}
	game.DealerChoice = gameReveal.DealerChoice
	return gameList, true
}

func GetGameByDealer(dealer string, gameList []Game) []Game {
	return gameList
}

func CloseGame(gameID uint32) {

}

func HandleGameContract(tx Transaction, gameList []Game) ([]Game, bool) {
	switch tx.ToAddress {
	case CreateGameAddress:
		fmt.Println("Create Game")
		gameList = CreateGame(tx, gameList)
	case JoinGameAddress:
		fmt.Println("Join Game")
		gameList = JoinGame(tx, gameList)
	case RevealChoiceAddress:
		fmt.Println("Reveal Game")
		return RevealGame(tx, gameList)
	default:
		return gameList, false
	}
	return gameList, true
}
