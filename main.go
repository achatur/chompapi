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
	"chompapi/crypto"
	"encoding/base64"
	"io/ioutil"
	"encoding/json"
)

type handler func(w http.ResponseWriter, r *http.Request)

func main() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/register", register.DoRegister)
	router.HandleFunc("/admin/fp", BasicAuth(register.ForgotPassword))
	router.HandleFunc("/login", login.DoLogin)
	router.HandleFunc("/me", me.GetMe)
	router.HandleFunc("/jwt", BasicAuth(crypto.GetJwt))
	router.HandleFunc("/me/photos", me.PostPhotoId)
	router.HandleFunc("/me/photos/{photoID}", me.PostPhotoId)
	router.HandleFunc("/me/reviews", me.Reviews)
	router.HandleFunc("/me/update/up", me.UpdatePassword)
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

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %v", html.EscapeString(r.URL.Path))
}

func init() {

	var err error
	GetConfig()

	globalsessionkeeper.GlobalSessions, err = session.NewManager("mysql", `{"EnableSetCookie":true, "Secure":true, "cookieLifeTime":604800, "CookieName":"chomp_sessionid","Gclifetime":300,"Maxlifetime":604800,"ProviderConfig":"root@tcp(172.16.0.1:3306)/chomp"}`)

	if err != nil {
		fmt.Printf("Error")
	}

	globalsessionkeeper.GlobalSessions.SetSecure(true)
	go globalsessionkeeper.GlobalSessions.GC()
}

func BasicAuth(pass handler) handler {
 
    return func(w http.ResponseWriter, r *http.Request) {

    	fmt.Println("made it to basic auth")
    	fmt.Printf("Headers = %v\n", r.Header)
 		fmt.Printf("Len = %v\n", len(r.Header))

 		if len(r.Header["Authorization"]) <= 0 {
 			http.Error(w, "bad syntax", http.StatusBadRequest)
			return
 		}
        auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)
 		fmt.Printf("auth = %v", auth)
        if len(auth) != 2 { 

            http.Error(w, "bad syntax", http.StatusBadRequest)
			return
        } else if auth[0] != "Basic" {
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
 
 func GetConfig() error {
	configFile, err := ioutil.ReadFile("./chomp_private/config.json")
	if err != nil {
	    return err
	}
	err = json.Unmarshal(configFile, &globalsessionkeeper.ChompConfig)
	if err != nil {
	    fmt.Printf("Err = %v", err)
	    return err
	}
	return nil
}

func Validate(username, password string) bool {
    fmt.Println("Made it to validate..")
    for _, e := range globalsessionkeeper.ChompConfig.Authorized  {
    	if e.User == username && e.Pass == password {
    		return true
    	}
    }

    return false
}
