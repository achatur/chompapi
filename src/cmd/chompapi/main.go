package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strings"
	"cmd/chompapi/login"
	// "cmd/chompapi/register"
	"cmd/chompapi/globalsessionkeeper"
	"github.com/astaxie/beego/session"
	// "cmd/chompapi/me"
	// "cmd/chompapi/review"
	"database/sql"
	_ "github.com/astaxie/beego/session/mysql"
	"github.com/gorilla/mux"
	"cmd/chompapi/crypto"
	"encoding/base64"
	"io/ioutil"
	"encoding/json"
	"cmd/chompapi/db"
	"reflect"
)

// type appContext struct {
//     db        *sql.DB
//     // store     *sessions.CookieStore
//     // templates map[string]*template.Template
//     // decoder   *schema.Decoder
//     // store     *redistore.RediStore
//     // mandrill  *gochimp.MandrillAPI
//     // twitter   *anaconda.TwitterApi
//     // log       *log.Logger
//     config      *globalsessionkeeper.ChompConfig // app-wide configuration: hostname, ports, etc.
//     // MyErrorResponse globalsessionkeeper.ErrorResponse
// }

type AppHandler struct { 
	*appContext
	h func(*appContext, http.ResponseWriter, *http.Request) (int, error)
}

func (ah globalsessionkeeper.AppContext) ServerHttp(w http.ResponseWriter, r *http.Request) {

	err := ah.h(ah.appContext, w, r)
	if err != nil {
		// log.Printf("HTTP %d: %q", status, err)
		switch status = err.(ErrorResponse).Code {
		case http.StatusNotFound:
			fmt.Printf("Error: Page not found\n")
			http.NotFound(w, r)
		case http.StatusInternalServerError:
			fmt.Printf("Error: %v\n", http.StatusInternalServerError)
			http.Error(w, http.StatusText(status), status)
		default:
			fmt.Printf("Error: %v\n", err)
			http.Error(w, http.StatusText(status), status)
		}
	}

}
// var MyErrorResponse globalsessionkeeper.ErrorResponse
var MyErrorResponse globalsessionkeeper.ErrorResponse
// var GlobalSessions *session.Manager

func main() {

	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	router := mux.NewRouter().StrictSlash(true)
	context := &appContext{db: db}

	router.HandleFunc("/login", AppHandler{context, login.DoLogin})

	port := "8000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	if strings.Contains(string(port), "443") {
		log.Fatal(http.ListenAndServeTLS(":"+port, globalsessionkeeper.ChompConfig.Cert.Cert, globalsessionkeeper.ChompConfig.Cert.Key, router))
	} else {
		log.Fatal(http.ListenAndServe(":" + port, router))
	}
}