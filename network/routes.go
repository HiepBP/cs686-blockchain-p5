package network

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"StartBlockChain",
		"GET",
		"/startbc",
		StartBlockChain,
	},
	Route{
		"Show",
		"GET",
		"/show",
		Show,
	},
	Route{
		"Upload",
		"GET",
		"/upload",
		Upload,
	},
	Route{
		"UploadBlock",
		"GET",
		"/block/{height}/{hash}",
		UploadBlock,
	},
	Route{
		"UploadPeer",
		"GET",
		"/peers/{address}/{id}",
		UploadPeer,
	},
	Route{
		"HeartBeatReceive",
		"POST",
		"/heartbeat/receive",
		HeartBeatReceive,
	},
	Route{
		"StartMiner",
		"GET",
		"/miner/start/{publicKey}",
		StartMiner,
	},
	Route{
		"Canonical",
		"GET",
		"/canonical",
		Canonical,
	},
	Route{
		"HandleTx",
		"POST",
		"/handleTx",
		HandleTx,
	},
	Route{
		"UploadAccounts",
		"GET",
		"/accounts/upload",
		UploadAccounts,
	},
	Route{
		"GetAccountBalaceFork",
		"GET",
		"/accounts/{publicKey}",
		GetAccountBalanceFork,
	},
	Route{
		"ShowBalance",
		"GET",
		"/accounts",
		ShowBalance,
	},
	Route{
		"GetGameInformation",
		"GET",
		"/games",
		GetGameInformation,
	},
}
