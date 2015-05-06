package main

import (
	"fmt"
	"github.com/astaxie/beego/session"
	_ "github.com/astaxie/beego/session/mysql"
	"html"
	"log"
	"net/http"
	"os"
	"strings"
)

//Global Variable
var globalSessions *session.Manager

func main() {
	http.HandleFunc("/register", doRegister)
	http.HandleFunc("/login", doLogin)

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
	//globalSessions, err = session.NewManager("mysql", `{"SessionOn":true, "cookieName":"gosessionid","gclifetime":3600,"ProviderConfig":"root:''@protocol(172.16.0.1:3306)/chomp"}`)
	globalSessions, err = session.NewManager("mysql", `{"enableSetCookie":true, "SessionOn":true, "cookieName":"chomp_sessionid","gclifetime":120,"ProviderConfig":"root@tcp(172.16.0.1:3306)/chomp"}`)
	if err != nil {
		fmt.Printf("Error")
	}
	globalSessions.SetSecure(true)
	go globalSessions.GC()
}
