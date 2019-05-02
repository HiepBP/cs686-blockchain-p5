package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"./wallet"
)

func main() {
	defer os.Exit(0)

	// if len(os.Args) > 2 {
	// 	router := web.NewRouter()
	// 	log.Fatal(http.ListenAndServe(":"+os.Args[2], router))
	// } else {
	// 	fmt.Println("Start BC\n")
	// 	router := network.NewRouter()
	// 	log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	// }

	privateKey, publicKey := wallet.GenerateKey()
	data := "Hello World"
	signature, err := wallet.SignData(data, hex.EncodeToString(privateKey))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	match, err := wallet.ValidateSignature(data, signature, hex.EncodeToString(publicKey))
	if match {
		fmt.Println("MATCH")
	} else {
		fmt.Println("NO MATCH")
	}
}
