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
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/register", register.DoRegister)
	router.HandleFunc("/login", login.DoLogin)
	router.HandleFunc("/me", me.GetMe)
	router.HandleFunc("/me/photos", me.PostPhotoId)
	router.HandleFunc("/me/photos/{photo_id}", me.PostPhotoId)
	//router.HandleFunc("/me/reviews", me.Reviews)
	router.HandleFunc("/reviews", review.Reviews)
	router.HandleFunc("/reviews/{review_id}", review.Reviews)

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
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func init() {
	var err error
	fmt.Println("Session init")
	globalsessionkeeper.GlobalSessions, err = session.NewManager("mysql", `{"EnableSetCookie":true, "Secure":true, "cookieLifeTime":604800, "CookieName":"chomp_sessionid","Gclifetime":604800,"Maxlifetime":604800,"ProviderConfig":"root@tcp(172.16.0.1:3306)/chomp"}`)
	fmt.Printf("started session, manager = %v\n", globalsessionkeeper.GlobalSessions)
	fmt.Printf("Type = %v\n", reflect.TypeOf(globalsessionkeeper.GlobalSessions))
	// fmt.Printf("Config = %v\n",globalsessionkeeper.GlobalSessions.provider)
	if err != nil {
		fmt.Printf("Error")
	}
	globalsessionkeeper.GlobalSessions.SetSecure(true)
	go globalsessionkeeper.GlobalSessions.GC()
}