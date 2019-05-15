package web

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
	"io/ioutil"

	"../network"
	"../network/data"
	"../wallet"
	"./helpers"
	"./models"
	"../blockchain"
	"golang.org/x/crypto/sha3"
)

var (
	CreateGameAddress   = "636f6e7472616374637265617465"
	JoinGameAddress     = "636f6e74726163746a6f696e"
	RevealChoiceAddress = "636f6e747261637472657665616c"
)

func StartBlockChain(w http.ResponseWriter, r *http.Request) {
	network.StartBlockChain(w, r)
	//Download blockchain and peerlist from host
	if "http://"+r.Host != network.FIRST_ADDR {
		network.SELF_ADDR = "http://" + r.Host
		network.Download()
	}
}

func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {
	network.HeartBeatReceive(w, r)
}

//GetLogin will return a html page to login
func GetLogin(w http.ResponseWriter, r *http.Request) {
	var body, _ = helpers.LoadFile("templates/login.html")
	fmt.Fprintf(w, body)
}

//PostLogin will handle logic to check username and password
func PostLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	publicKey := r.FormValue("publicKey")
	redirectTarget := "/login"
	if !helpers.IsEmpty(publicKey) {
		_, portStr, _ := net.SplitHostPort(r.Host)
		portNumber, _ := strconv.Atoi(portStr)
		// Check if account is added before
		network.SBC.GetLatestBlocks()
		downPeerList(int32(portNumber))
		helpers.SetCookie(publicKey, w)
		redirectTarget = "/"
	}
	http.Redirect(w, r, redirectTarget, 302)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	helpers.ClearCookie(w)
	http.Redirect(w, r, "/login", 302)
}

func Index(w http.ResponseWriter, r *http.Request) {
	publicKey := helpers.CheckLogin(r)
	if !helpers.IsEmpty(publicKey) {
		account := getAccount(publicKey)
		balance := account.Balance
		var indexBody, _ = helpers.LoadFile("templates/index.html")
		fmt.Fprintf(w, indexBody, publicKey, balance)
	} else {
		http.Redirect(w, r, "/login", 302)
	}
}

func Show(w http.ResponseWriter, r *http.Request) {
	network.Show(w, r)
}

func CreateGame(w http.ResponseWriter, r *http.Request) {
	publicKey := helpers.CheckLogin(r)
	if helpers.IsEmpty(publicKey) {
		http.Redirect(w, r, "/login", 302)
	} else {
		r.ParseForm()
		choice := r.FormValue("choice")
		gameValue, _ := strconv.Atoi(r.FormValue("gameValue"))
		secretNumber := r.FormValue("secretNumber")
		privateKey := r.FormValue("privateKey")
		fee, _ := strconv.Atoi(r.FormValue("fee"))

		dealerHash := sha3.Sum256([]byte(choice + secretNumber))
		game := models.GameCreate{
			DealerHash: hex.EncodeToString(dealerHash[:]),
		}
		gameJSON, _ := game.EncodeToJSON()
		transaction := data.Transaction{
			FromAddress: publicKey,
			ToAddress:   CreateGameAddress,
			TimeStamp:   time.Now().Unix(),
			Value:       gameValue,
			Data:        gameJSON,
			Fee: fee,
		}
		//Fake transaction with timestamp
		transaction.ID = hex.EncodeToString(transaction.Hash())
		txJSON, _ := transaction.EncodeToJSON()
		signature, _ := wallet.SignData(txJSON, privateKey)
		valid, _ := wallet.ValidateSignature(txJSON, signature, publicKey)
		if !valid {
			fmt.Fprintln(w, "Invalid key")
			return
		}
		fmt.Println(txJSON)
		signedTx := data.SignedTransaction{
			Transaction: transaction,
			Signature:   signature,
		}
		go sendTx(signedTx)
		var body, _ = helpers.LoadFile("templates/info.html")
		fmt.Fprintf(w, body, "Create Game")
	}
}

func JoinGame(w http.ResponseWriter, r *http.Request) {
	publicKey := helpers.CheckLogin(r)
	if helpers.IsEmpty(publicKey) {
		http.Redirect(w, r, "/login", 302)
	} else {
		r.ParseForm()
		choice, _ := strconv.Atoi(r.FormValue("choice"))
		gameValue, _ := strconv.Atoi(r.FormValue("gameValue"))
		id, _ := strconv.Atoi(r.FormValue("id"))
		privateKey := r.FormValue("privateKey")
		fee, _ := strconv.Atoi(r.FormValue("fee"))
		game := models.GameJoin{
			ID:           uint32(id),
			PlayerChoice: uint32(choice),
		}
		gameJSON, _ := game.EncodeToJSON()
		transaction := data.Transaction{
			FromAddress: publicKey,
			ToAddress:   JoinGameAddress,
			TimeStamp:   time.Now().Unix(),
			Value:       gameValue,
			Data:        gameJSON,
			Fee: fee,
		}
		transaction.ID = hex.EncodeToString(transaction.Hash())
		txJSON, _ := transaction.EncodeToJSON()
		signature, _ := wallet.SignData(txJSON, privateKey)
		valid, _ := wallet.ValidateSignature(txJSON, signature, publicKey)
		if !valid {
			fmt.Fprintln(w, "Invalid key")
			return
		}
		fmt.Println(txJSON)
		signedTx := data.SignedTransaction{
			Transaction: transaction,
			Signature:   signature,
		}
		go sendTx(signedTx)
		var body, _ = helpers.LoadFile("templates/info.html")
		fmt.Fprintf(w, body, "Join Game")
	}
}

func RevealGame(w http.ResponseWriter, r *http.Request) {
	publicKey := helpers.CheckLogin(r)
	if helpers.IsEmpty(publicKey) {
		http.Redirect(w, r, "/login", 302)
	} else {
		r.ParseForm()
		choice, _ := strconv.Atoi(r.FormValue("choice"))
		gameValue, _ := strconv.Atoi(r.FormValue("gameValue"))
		secretNumber := r.FormValue("secretNumber")
		id, _ := strconv.Atoi(r.FormValue("id"))
		privateKey := r.FormValue("privateKey")
		fee, _ := strconv.Atoi(r.FormValue("fee"))
		game := models.GameReveal{
			ID:           uint32(id),
			DealerChoice: uint32(choice),
			SecretNumber: secretNumber,
		}
		gameJSON, _ := game.EncodeToJSON()
		transaction := data.Transaction{
			FromAddress: publicKey,
			ToAddress:   RevealChoiceAddress,
			TimeStamp:   time.Now().Unix(),
			Value:       gameValue,
			Data:        gameJSON,
			Fee: fee,
		}
		transaction.ID = hex.EncodeToString(transaction.Hash())
		txJSON, _ := transaction.EncodeToJSON()
		signature, _ := wallet.SignData(txJSON, privateKey)
		valid, _ := wallet.ValidateSignature(txJSON, signature, publicKey)
		if !valid {
			fmt.Fprintln(w, "Invalid key")
			return
		}
		fmt.Println(txJSON)
		signedTx := data.SignedTransaction{
			Transaction: transaction,
			Signature:   signature,
		}
		go sendTx(signedTx)
		var body, _ = helpers.LoadFile("templates/info.html")
		fmt.Fprintf(w, body, "Reveal Game")
	}
}

func Games(w http.ResponseWriter, r *http.Request){
	publicKey := helpers.CheckLogin(r)
	if helpers.IsEmpty(publicKey) {
		http.Redirect(w, r, "/login", 302)
	} else{gameAccount := getAccount(network.ContractAddress)
		games, _ := data.DecodeGameListFromJSON(gameAccount.Data)
		for _, game := range games{
			dealerChoice := "?"
			playerChoice := "?"
			switch game.DealerChoice {
			case 1:
				dealerChoice = "rock"
			case 2:
				dealerChoice = "paper"
			case 3:
				dealerChoice = "scissors"	
			}
			if game.Player==""{
				fmt.Fprintf(w, "ID: %d - Bet: %d \nDealer: %s - Dealer Choice: %s\n\n",
					game.ID, game.DealerValue, game.Dealer[:5], dealerChoice)
			} else {
				switch game.PlayerChoice {
				case 1:
					playerChoice = "rock"
				case 2:
					playerChoice = "paper"
				case 3:
					playerChoice = "scissors"	
				}
				if game.DealerChoice==0{
					fmt.Fprintf(w, "ID: %d - Bet: %d \nDealer: %s - Dealer Choice: %s \nPlayer: %s - Player Choice: %s\n\n",
						game.ID, game.DealerValue, game.Dealer[:5], dealerChoice, game.Player[:5], playerChoice)
				} else {
					result := "Draw"
					switch game.Result{
					case 201:
						result = "Dealer Win"
					case 102:
						result = "Player Win"
					}
					fmt.Fprintf(w, "ID: %d - Bet: %d - Result: %s \nDealer: %s - Dealer Choice: %s \nPlayer: %s - Player Choice: %s\n\n",
						game.ID, game.DealerValue, result, game.Dealer[:5], dealerChoice, game.Player[:5], playerChoice)
				}
			}
		}
	}
	
}

func sendTx(signedTx data.SignedTransaction) {
	signedTxJSON, _ := signedTx.EncodeToJSON()
	peers := network.Peers.Copy()
	for addr := range peers {
		currAddr := addr
		go func() {
			_, err := http.Post(currAddr+"/handleTx", "application/json", bytes.NewBuffer([]byte(signedTxJSON)))
			if err != nil {
				fmt.Println("Peer not available")
				network.Peers.Delete(currAddr)
			}
		}()

	}
}

func downPeerList(portNumber int32) {
	network.RegisterLocal(portNumber)
	network.DownloadPeerList()
}

func getAccount(publicKey string) bc.Account{
	peers := network.Peers.Copy()
	for addr := range peers {
		response, err := http.Get(addr + "/accounts/" + publicKey)
		if err != nil {
			fmt.Println("Peer not available")
			network.Peers.Delete(addr)
		} else {
			accountJSONBuffer, _ := ioutil.ReadAll(response.Body)
			accountJSON := string(accountJSONBuffer)
			account,_ := bc.DecodeAccountFromJSON(accountJSON)
			return account
		}

	}
	return bc.Account{}
}