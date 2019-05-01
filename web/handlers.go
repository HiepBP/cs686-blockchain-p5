package web

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"../network"
	"../wallet"
	"./helpers"
	"./models"
	"golang.org/x/crypto/sha3"
)

var (
	accounts []models.Account
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

//GetRegister will return a html page to create account
func GetRegister(w http.ResponseWriter, r *http.Request) {
	var body, _ = helpers.LoadFile("templates/register.html")
	fmt.Fprintf(w, body)
}

//PostRegister will handle logic to create account
func PostRegister(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	_uName, _pwd := false, false
	_uName = !helpers.IsEmpty(username)
	_pwd = !helpers.IsEmpty(password)

	if _uName && _pwd {
		for _, account := range accounts {
			if account.Username == username {
				fmt.Fprintln(w, "This username already created!")
			}
		}
		privateKey, publicKey := wallet.GenerateKey()
		newAccount := models.Account{
			Username:  username,
			Password:  password,
			PublicKey: publicKey,
		}
		accounts = append(accounts, newAccount)
		accountJSON, _ := newAccount.EncodeToJSON()
		network.SendNewAccount(accountJSON)
		fmt.Fprintln(w, "Username for Register : ", username)
		fmt.Fprintln(w, "Password for Register : ", password)
		fmt.Fprintln(w, "Your private key, save it: ", string(privateKey))
	} else {
		fmt.Fprintln(w, "Username for Register : ", username)
		fmt.Fprintln(w, "Password for Register : ", password)
		fmt.Fprintln(w, "This fields can not be blank!")
	}
}

//GetLogin will return a html page to login
func GetLogin(w http.ResponseWriter, r *http.Request) {
	var body, _ = helpers.LoadFile("templates/login.html")
	fmt.Fprintf(w, body)
}

//PostLogin will handle logic to check username and password
func PostLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	fmt.Println(username)
	fmt.Println(password)
	redirectTarget := "/login"
	if !helpers.IsEmpty(username) && !helpers.IsEmpty(password) {
		// Database check for user data!
		for _, account := range accounts {
			if account.CheckAccount(username, password) {
				helpers.SetCookie(username, string(account.PublicKey), w)
				redirectTarget = "/"
			}
		}
		if redirectTarget == "/login" {
			redirectTarget = "/register"
		}
	}
	http.Redirect(w, r, redirectTarget, 302)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	helpers.ClearCookie(w)
	http.Redirect(w, r, "/login", 302)
}

func Index(w http.ResponseWriter, r *http.Request) {
	username, _ := helpers.CheckLogin(r)
	if !helpers.IsEmpty(username) {
		var indexBody, _ = helpers.LoadFile("templates/index.html")
		fmt.Fprintf(w, indexBody, username)
	} else {
		http.Redirect(w, r, "/login", 302)
	}
}

func CreateGame(w http.ResponseWriter, r *http.Request) {
	username, publicKey := helpers.CheckLogin(r)
	if helpers.IsEmpty(username) {
		http.Redirect(w, r, "/login", 302)
	} else {
		r.ParseForm()
		choice := r.FormValue("choice")
		gameValue, _ := strconv.ParseFloat(r.FormValue("gameValue"), 32)
		privateKey := r.FormValue("key")
		dealerHash := sha3.Sum256([]byte(choice + privateKey))
		game := models.Game{
			ID:           1,
			Dealer:       publicKey,
			DealerChoice: 0,
			DealerHash:   hex.EncodeToString(dealerHash[:]),
			Player:       "",
			PlayerChoice: 0,
			GameValue:    float32(gameValue),
			Result:       0,
			Closed:       false,
		}
		fmt.Println(w, game)
	}
}

func Show(w http.ResponseWriter, r *http.Request) {
	network.Show(w, r)
}
