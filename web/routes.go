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
		"Register",
		"GET",
		"/register",
		GetRegister,
	},
	Route{
		"Register",
		"POST",
		"/register",
		PostRegister,
	},
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
}
