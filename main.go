package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"./network"
	"./web"
)

func main() {
	defer os.Exit(0)

	if len(os.Args) > 2 {
		router := web.NewRouter()
		log.Fatal(http.ListenAndServe(":"+os.Args[2], router))
	} else {
		fmt.Println("Start BC\n")
		router := network.NewRouter()
		log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	}

	// privateKey := "1bbfc1c4289e9081e153671bb98c6564fc4a29dfe63b1a62cbf93056a3e56c09"
	// publicKey := "0474b1a963385c934234f20a987e50a3b85ffe20777b23ca8fb0fc88768120bae96a104fdfb774005517c03a0e9cacb7761901a57c5beb0b2895dcd4d7f0fadc94"
	// data := "{\"dealerHash\":\"7a5df5ffa0dec2228d90b8d0a0f1b0767b748b0a41314c123075b8289e4e053f\",\"gameValue\":\"10\"}"
	// fmt.Println(data)
	// signature, err := wallet.SignData(data, privateKey)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// // signatureStr := "3046022100d9db98867c7a5d48d184f53f6df0d06cec6784d2dad9cb5cf2cb6df28bd4ff28022100931753da7d2d74aea04c1063e8995f993ab668476b608ece152581ca31341e16"
	// // signature, _ := hex.DecodeString(signatureStr)
	// fmt.Println(signature)
	// fmt.Println(hex.EncodeToString(signature))
	// match, _ := wallet.ValidateSignature(data, signature, publicKey)
	// if match {
	// 	fmt.Println("MATCH")
	// } else {
	// 	fmt.Println("NO MATCH")
	// }
}
