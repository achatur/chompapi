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
)

func main() {
	http.HandleFunc("/register", register.DoRegister)
	http.HandleFunc("/login", login.DoLogin)
	http.HandleFunc("/me", me.GetMe)

	port := os.Getenv("PORT")
	if strings.Contains(string(port), "443") {
		log.Fatal(http.ListenAndServeTLS(":"+port, "/home/amir.chatur/working/playground/gen_cert/thechompapp.com.pem", "/home/amir.chatur/working/playground/gen_cert/thechompapp.com.key.pem", nil))
	} else {
		log.Fatal(http.ListenAndServe(":8000", nil))
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func init() {
	var err error
	fmt.Println("Session init")
	globalsessionkeeper.GlobalSessions, err = session.NewManager("mysql", `{"enableSetCookie":true, "SessionOn":true, "cookieName":"chomp_sessionid","gclifetime":300,"ProviderConfig":"root@tcp(172.16.0.1:3306)/chomp"}`)
	if err != nil {
		fmt.Printf("Error")
	}
	globalsessionkeeper.GlobalSessions.SetSecure(true)
	go globalsessionkeeper.GlobalSessions.GC()
}