package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strings"
	"chompapi/login"
	"chompapi/register"
	"chompapi/globalsessionkeeper"
	"github.com/astaxie/beego/session"
	"chompapi/me"
	"chompapi/review"
	"github.com/gorilla/mux"
	"reflect"
	"chompapi/crypto"
)

func main() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/register", register.DoRegister)
	router.HandleFunc("/login", login.DoLogin)
	router.HandleFunc("/me", me.GetMe)
	router.HandleFunc("/jwt", BasicAuth(crypto.GetJwt)
	router.HandleFunc("/me/photos", me.PostPhotoId)
	router.HandleFunc("/me/photos/{photoID}", me.PostPhotoId)
	router.HandleFunc("/me/reviews", me.Reviews)
	router.HandleFunc("/reviews", review.Reviews)
	router.HandleFunc("/reviews/{reviewID}", review.Reviews)

	port := "8000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	if strings.Contains(string(port), "443") {
		log.Fatal(http.ListenAndServeTLS(":"+port, "/home/amir.chatur/working/playground/gen_cert/thechompapp.com.pem", "/home/amir.chatur/working/playground/gen_cert/thechompapp.com.key.pem", router))
	} else {
		log.Fatal(http.ListenAndServe(":" + port, router))
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %v", html.EscapeString(r.URL.Path))
}

func init() {

	var err error

	globalsessionkeeper.GlobalSessions, err = session.NewManager("mysql", `{"EnableSetCookie":true, "Secure":true, "cookieLifeTime":604800, "CookieName":"chomp_sessionid","Gclifetime":300,"Maxlifetime":604800,"ProviderConfig":"root@tcp(172.16.0.1:3306)/chomp"}`)

	if err != nil {
		fmt.Printf("Error")
	}

	globalsessionkeeper.GlobalSessions.SetSecure(true)
	go globalsessionkeeper.GlobalSessions.GC()
}

func BasicAuth(pass handler) handler {
 
    return func(w http.ResponseWriter, r *http.Request) {
 
        auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)
 
        if len(auth) != 2 || auth[0] != "Basic" {
            http.Error(w, "bad syntax", http.StatusBadRequest)
            return
        }
 
        payload, _ := base64.StdEncoding.DecodeString(auth[1])
        pair := strings.SplitN(string(payload), ":", 2)
 
        if len(pair) != 2 || !Validate(pair[0], pair[1]) {
            http.Error(w, "authorization failed", http.StatusUnauthorized)
            return
        }
 
        pass(w, r)
    }
}
 
func Validate(username, password string) bool {
    if username == "username" && password == "password" {
        return true
    }
    return false
}
