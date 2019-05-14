package web

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
		"Login",
		"GET",
		"/login",
		GetLogin,
	},
	Route{
		"Login",
		"POST",
		"/login",
		PostLogin,
	},
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},
	Route{
		"Logout",
		"POST",
		"/logout",
		Logout,
	},
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
		"CreateGame",
		"POST",
		"/games/create",
		CreateGame,
	},
	Route{
		"JoinGame",
		"POST",
		"/games/join",
		JoinGame,
	},
	Route{
		"RevealGame",
		"POST",
		"/games/reveal",
		RevealGame,
	},
	Route{
		"Games",
		"GET",
		"/games",
		Games,
	},
}
