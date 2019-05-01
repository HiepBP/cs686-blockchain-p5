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

	// cmd := cli.CommandLine{}
	// cmd.Run()
	if len(os.Args) > 2 {
		router := web.NewRouter()
		log.Fatal(http.ListenAndServe(":"+os.Args[2], router))
	} else {
		fmt.Println("Start BC\n")
		router := network.NewRouter()
		log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	}
	// p3.Register()
	// data.TestPeerListRebalance()
}
