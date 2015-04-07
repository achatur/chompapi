package main

import (
	"fmt"
	"github.com/astaxie/beego/session"
	//_ "github.com/astaxie/session/providers/memory"
	//"github.com/astaxie/beego/session/providers/mysql"
	_ "github.com/astaxie/beego/session/mysql"
	//"github.com/go-sql-driver/mysql"
	"html"
	"log"
	"net/http"
	//"session"
)

var globalSessions *session.Manager

//var mysqlProvider = &session.MysqlProvider

func main() {
	http.HandleFunc("/register", doRegister)
	http.HandleFunc("/login", doLogin)

	log.Fatal(http.ListenAndServeTLS(":8443", "/home/amir.chatur/working/playground/gen_cert/thechompapp.com.pem", "/home/amir.chatur/working/playground/gen_cert/thechompapp.com.key.pem", nil))
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
