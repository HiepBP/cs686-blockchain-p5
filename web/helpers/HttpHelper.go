package helpers

import "net/http"
import "github.com/gorilla/securecookie"

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func SetCookie(username string, publicKey string, response http.ResponseWriter) {
	value := map[string]string{
		"username":  username,
		"publicKey": publicKey,
	}
	if encoded, err := cookieHandler.Encode("cookie", value); err == nil {
		cookie := &http.Cookie{
			Name:  "cookie",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func ClearCookie(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "cookie",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

func CheckLogin(request *http.Request) (username string, publicKey string) {
	if cookie, err := request.Cookie("cookie"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("cookie", cookie.Value, &cookieValue); err == nil {
			username = cookieValue["username"]
			publicKey = cookieValue["publicKey"]
		}
	}
	return username, publicKey
}
