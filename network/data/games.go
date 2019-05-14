package data

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	bc "../../blockchain"
	"../../mpt"
	"golang.org/x/crypto/sha3"
)

var (
	DEALERWIN           uint32 = 201
	PLAYERWIN           uint32 = 102
	DRAW                uint32 = 101
	CreateGameAddress          = "636f6e7472616374637265617465"
	JoinGameAddress            = "636f6e74726163746a6f696e"
	RevealChoiceAddress        = "636f6e747261637472657665616c"
	GAMERESULT                 = [][]uint32{ //1 rock, 2 paper, 3 scissors
		{DRAW, PLAYERWIN, PLAYERWIN, PLAYERWIN},
		{DEALERWIN, DRAW, PLAYERWIN, DEALERWIN},
		{DEALERWIN, DEALERWIN, DRAW, PLAYERWIN},
		{DEALERWIN, PLAYERWIN, DEALERWIN, DRAW},
	}
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
	game.DealerValue = tx.Value
	gameList = append(gameList, game)

	return gameList
}

func JoinGame(tx Transaction, gameList []Game) ([]Game, bool) {
	gameJoin, _ := DecodeGameJoinFromJSON(tx.Data)
	if int(gameJoin.ID) >= len(gameList) {
		return gameList, false
	}
	game := gameList[gameJoin.ID]

	if game.Dealer == tx.FromAddress || game.DealerValue!=tx.Value || game.PlayerChoice!=0 || 
		game.Closed || !checkChoice(gameJoin.PlayerChoice){
		fmt.Println("Something wrong when join game")
		return gameList, false
	}

	game.Player = tx.FromAddress
	game.PlayerChoice = gameJoin.PlayerChoice
	game.PlayerValue = tx.Value
	gameList[gameJoin.ID] = game
	return gameList, true
}

func RevealGame(tx Transaction, gameAccount *bc.Account, accountTrie *mpt.MerklePatriciaTrie) bool {
	gameList, _ := DecodeGameListFromJSON(gameAccount.Data)
	gameReveal, _ := DecodeGameRevealFromJSON(tx.Data)
	game := gameList[gameReveal.ID]

	if game.Dealer != tx.FromAddress || game.PlayerChoice==0 || 
		game.Closed || !checkChoice(gameReveal.DealerChoice) || !checkChoice(game.PlayerChoice){
		fmt.Println("Game cannot be revealed now")
		return false
	}

	dealerHash := sha3.Sum256([]byte(strconv.Itoa(int(gameReveal.DealerChoice)) + gameReveal.SecretNumber))
	if hex.EncodeToString(dealerHash[:]) != game.DealerHash {
		fmt.Println("Incorrect secret number")
		return false
	}
	game.DealerChoice = gameReveal.DealerChoice
	gameList[gameReveal.ID] = game
	gameListJSON, _ := EncodeGameListToJSON(gameList)
	gameAccount.Data = gameListJSON
	CloseGame(gameReveal.ID, gameAccount, accountTrie)
	return true
}

func CloseGame(gameID uint32, gameAccount *bc.Account, accountTrie *mpt.MerklePatriciaTrie) bool {
	gameList, _ := DecodeGameListFromJSON(gameAccount.Data)
	game := gameList[gameID]
	if game.Closed || (game.DealerChoice == 0 && game.PlayerChoice == 0) {
		return false
	}
	result := GAMERESULT[game.DealerChoice][game.PlayerChoice]
	//Do transfer money
	if result == DEALERWIN {
		fmt.Println("DEALER WIN")
		transfer(gameAccount, game.Dealer, game.DealerValue + game.PlayerValue, accountTrie)
	} else if result == PLAYERWIN {
		fmt.Println("PLAYER WIN")
		transfer(gameAccount, game.Player, game.DealerValue + game.PlayerValue, accountTrie)
	} else if result == DRAW {
		fmt.Println("DRAW WIN")
		transfer(gameAccount, game.Dealer, game.DealerValue, accountTrie)
		transfer(gameAccount, game.Player, game.PlayerValue, accountTrie)
	}
	game.Closed = true
	game.Result = result
	gameList[gameID] = game
	gameListJSON, _ := EncodeGameListToJSON(gameList)
	gameAccount.Data = gameListJSON
	fmt.Println("Finish Reveal Game")
	return true
}

func GetGameByDealer(dealer string, gameList []Game) []Game {
	var result []Game
	for _, game := range gameList {
		if game.Dealer == dealer {
			result = append(result, game)
		}
	}
	return result
}

func HandleGameContract(tx Transaction, accountTrie *mpt.MerklePatriciaTrie) bool {
	gameAccountJSON, _ := accountTrie.Get(ContractAddress)
	gameAccount, _ := bc.DecodeAccountFromJSON(gameAccountJSON)
	gameList, _ := DecodeGameListFromJSON(gameAccount.Data)
	ok := false
	switch tx.ToAddress {
	case CreateGameAddress:
		fmt.Println("Create Game")
		gameList = CreateGame(tx, gameList)
		gameListJSON, _ := EncodeGameListToJSON(gameList)
		gameAccount.Data = gameListJSON
		ok = true
	case JoinGameAddress:
		fmt.Println("Join Game")
		gameList, ok = JoinGame(tx, gameList)
		if ok {
			gameListJSON, _ := EncodeGameListToJSON(gameList)
			gameAccount.Data = gameListJSON
		}
	case RevealChoiceAddress:
		fmt.Println("Reveal Game")
		ok = RevealGame(tx, &gameAccount, accountTrie)
	}
	if ok {
		gameAccountJSON, _ := gameAccount.EncodeToJSON()
		//Change function address to contract address to update balance
		tx.ToAddress = ContractAddress
		accountTrie.Insert(tx.ToAddress, gameAccountJSON)
	}
	return ok
}

func transfer(gameAccount *bc.Account, to string, value int, accountTrie *mpt.MerklePatriciaTrie) {
	ToAccountJSON, _ := accountTrie.Get(to)
	ToAccount, _ := bc.DecodeAccountFromJSON(ToAccountJSON)
	gameAccount.Balance = gameAccount.Balance - value
	ToAccount.Balance = ToAccount.Balance + value
	ToAccountJSON, _ = ToAccount.EncodeToJSON()
	fmt.Println("Transfer: ", ToAccountJSON)
	accountTrie.Insert(to, ToAccountJSON)
}

func checkChoice(choice uint32) bool {
	return choice == 1 || choice ==2 || choice==3
}