package web

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"../network"
	"../network/data"
	"../wallet"
	"./helpers"
	"./models"
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
		var indexBody, _ = helpers.LoadFile("templates/index.html")
		fmt.Fprintf(w, indexBody, publicKey)
	} else {
		http.Redirect(w, r, "/login", 302)
	}
}

func CreateGame(w http.ResponseWriter, r *http.Request) {
	publicKey := helpers.CheckLogin(r)
	if helpers.IsEmpty(publicKey) {
		http.Redirect(w, r, "/login", 302)
	} else {
		r.ParseForm()
		choice := r.FormValue("choice")
		gameValue, _ := strconv.ParseFloat(r.FormValue("gameValue"), 32)
		secretNumber := r.FormValue("secretNumber")
		privateKey := r.FormValue("privateKey")

		dealerHash := sha3.Sum256([]byte(choice + secretNumber))
		game := models.Game{
			ID:           1,
			Dealer:       publicKey,
			DealerChoice: 0,
			DealerHash:   hex.EncodeToString(dealerHash[:]),
			Player:       "",
			PlayerChoice: 0,
			GameValue:    int(gameValue),
			Result:       0,
			Closed:       false,
		}
		// fmt.Println(publicKey)
		// fmt.Println(privateKey)
		gameJSON, _ := game.EncodeToJSON()
		// data := models.UserMsg{
		// 	Signature: "CreateGame",
		// 	Data:      gameJSON,
		// }
		// dataJSON, _ := data.EncodeToJSON()
		transaction := data.Transaction{
			FromAddress: publicKey,
			ToAddress:   CreateGameAddress,
			Value:       game.GameValue,
			Data:        gameJSON,
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
		sendTx(signedTx)
		fmt.Fprintf(w, "Please wait for your game to be created")
	}
}

func Show(w http.ResponseWriter, r *http.Request) {
	network.Show(w, r)
}

func sendTx(signedTx data.SignedTransaction) {
	signedTxJSON, _ := signedTx.EncodeToJSON()
	peers := network.Peers.Copy()
	for addr := range peers {
		_, err := http.Post(addr+"/handleTx", "application/json", bytes.NewBuffer([]byte(signedTxJSON)))
		if err != nil {
			fmt.Println("Peer not available")
			network.Peers.Delete(addr)
		}
	}
}

func downPeerList(portNumber int32) {
	network.RegisterLocal(portNumber)
	network.DownloadPeerList()
}
