package main

import (
	"fmt"
	"encoding/json"

  "time"
  "io/ioutil"
	"html"
	"log"
	"net/http"
	"os"
	"strings"
	"chompapi/login"
	"chompapi/register"
	"chompapi/globalsessionkeeper"
	"github.com/astaxie/beego/session"
	"github.com/dgrijalva/jwt-go"
	"chompapi/me"
	"chompapi/review"
	"github.com/gorilla/mux"
	"reflect"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/register", register.DoRegister)
	router.HandleFunc("/login", login.DoLogin)
	router.HandleFunc("/me", me.GetMe)
	router.HandleFunc("/jwt", GetJwt)
	router.HandleFunc("/me/photos", me.PostPhotoId)
	router.HandleFunc("/me/photos/{photoID}", me.PostPhotoId)
	router.HandleFunc("/me/reviews", me.Reviews)
	router.HandleFunc("/reviews", review.Reviews)
	router.HandleFunc("/reviews/{reviewID}", review.Reviews)
	router.HandleFunc("/insta/crawl", review.Crawl)

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

type JWT struct {
  JWT string `json:"jwt"`
}

func GetJwt(w http.ResponseWriter, r *http.Request) {
    token := jwt.New(jwt.SigningMethodRS256)
    mySigningKey, _ := ioutil.ReadFile("./test_key")
    // Set some claims

    token.Claims["scope"] = `https://www.googleapis.com/auth/devstorage.full_control`
    token.Claims["iss"] = "486543155383-oo5gldbn5q9jm3mei3de3p5p95ffn8fi@developer.gserviceaccount.com"
    token.Claims["iat"] = time.Now().Unix()
    token.Claims["exp"] = time.Now().Add(time.Hour * 1).Unix()
    token.Claims["aud"] = `https://www.googleapis.com/oauth2/v3/token`
    fmt.Println("%v", token.Claims)
    // Sign and get the complete encoded token as a string
    tokenString, _ := token.SignedString(mySigningKey)
    jwt := JWT{tokenString}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(jwt)
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func init() {
	var err error
	fmt.Println("Session init")
	globalsessionkeeper.GlobalSessions, err = session.NewManager("mysql", `{"EnableSetCookie":true, "Secure":true, "cookieLifeTime":604800, "CookieName":"chomp_sessionid","Gclifetime":300,"Maxlifetime":604800,"ProviderConfig":"root@tcp(172.16.0.1:3306)/chomp"}`)
	fmt.Printf("started session, manager = %v\n", globalsessionkeeper.GlobalSessions)
	fmt.Printf("Type = %v\n", reflect.TypeOf(globalsessionkeeper.GlobalSessions))
	// fmt.Printf("Config = %v\n",globalsessionkeeper.GlobalSessions.provider)
	if err != nil {
		fmt.Printf("Error")
	}
	globalsessionkeeper.GlobalSessions.SetSecure(true)
	go globalsessionkeeper.GlobalSessions.GC()
}
