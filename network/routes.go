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
		"/miner/start",
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
		"NewAccountReceive",
		"POST",
		"/accounts/receive",
		NewAccountReceive,
	},
}
