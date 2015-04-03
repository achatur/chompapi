package main

import (
	"fmt"
	"github.com/astaxie/beego/session"
	_ "github.com/astaxie/session/providers/memory"
	"html"
	"log"
	"net/http"
	//"session"
)

var globalSessions *session.Manager

func main() {
	http.HandleFunc("/register", doRegister)
	http.HandleFunc("/login", doLogin)

	//	log.Fatal(http.ListenAndServe(":8080", nil))
	log.Fatal(http.ListenAndServeTLS(":8443", "/home/amir.chatur/working/playground/gen_cert/thechompapp.com.pem", "/home/amir.chatur/working/playground/gen_cert/thechompapp.com.key.pem", nil))
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func init() {
	var err error
	globalSessions, err = session.NewManager("memory", `{"cookieName":"gosessionid","gclifetime":3600}`)
	if err != nil {
		fmt.Printf("Error")
	}
	go globalSessions.GC()
}
