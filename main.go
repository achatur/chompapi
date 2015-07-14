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
	_ "github.com/astaxie/beego/session/mysql"
	"chompapi/me"
	"chompapi/review"
	"github.com/gorilla/mux"
	"chompapi/crypto"
	"encoding/base64"
	"io/ioutil"
	"encoding/json"
	"chompapi/db"
	"reflect"
)

type handler func(w http.ResponseWriter, r *http.Request)
var MyErrorResponse globalsessionkeeper.ErrorResponse
var GlobalSessions *session.Manager

func main() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/register", register.DoRegister)
	router.HandleFunc("/login", login.DoLogin)

	router.HandleFunc("/admin/fp", BasicAuth(register.ForgotPassword))
	router.HandleFunc("/admin/fu", BasicAuth(register.ForgotUsername))
	router.HandleFunc("/admin/jwt", BasicAuth(crypto.GetJwt))

	router.HandleFunc("/me", SessionAuth(me.GetMe))
	router.HandleFunc("/me/logout", SessionAuth(me.Logout))
	router.HandleFunc("/me/logout/all", SessionAuth(me.LogoutAll))
	router.HandleFunc("/me/photos", SessionAuth(me.PostPhotoId))
	router.HandleFunc("/me/photos/{photoID}", SessionAuth(me.PostPhotoId))
	router.HandleFunc("/me/reviews", SessionAuth(me.Reviews))
	router.HandleFunc("/me/update/up", SessionAuth(me.UpdatePassword))
	router.HandleFunc("/me/update/d/{userID}", SessionAuth(me.DeleteMe))
	router.HandleFunc("/me/update/da/{userID}", SessionAuth(me.DeactivateMe))


	router.HandleFunc("/reviews", review.Reviews)
	router.HandleFunc("/reviews/{reviewID}", review.Reviews)
	
	router.HandleFunc("/insta/crawl", review.Crawl)

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

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %v", html.EscapeString(r.URL.Path))
}

func init() {

	var err error
	GetConfig()
	sessionConfig, _ := json.Marshal(globalsessionkeeper.ChompConfig.ManagerConfig)
	globalsessionkeeper.GlobalSessions, err = session.NewManager("mysql", string(sessionConfig))

	if err != nil {
		fmt.Printf("Coud not start session..Error: %v\n", err.Error())
		os.Exit(-1)

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

func SessionAuth(pass handler) handler {

	return func(w http.ResponseWriter, r *http.Request) {
		cookie := globalsessionkeeper.GetCookie(r)
		if cookie == "" {
			//need logging here instead of print
			fmt.Println("Cookie = %v", cookie)
			MyErrorResponse.Code = http.StatusUnauthorized
			MyErrorResponse.Error = "No Cookie Present"
			MyErrorResponse.HttpErrorResponder(w)
			return
		}
	
		sessionStore, err := globalsessionkeeper.GlobalSessions.GetSessionStore(cookie)
		if err != nil {
			//need logging here instead of print
			MyErrorResponse.Code = http.StatusUnauthorized
			MyErrorResponse.Error = "Session Expired"
			MyErrorResponse.HttpErrorResponder(w)
			return
		}
	
		sessionUser := sessionStore.Get("username")
		fmt.Println("SessionUser = %v", sessionUser)
		if sessionUser == nil {
			//need logging here instead of print
			fmt.Printf("Username not found, returning unauth, Get has %v\n", sessionStore)
			MyErrorResponse.Code = http.StatusUnauthorized
			MyErrorResponse.Error = "Session Expired"
			MyErrorResponse.HttpErrorResponder(w)
			return
		}

		fmt.Println("Getting user info for user %v\n", sessionUser)
		userInfo := new(db.UserInfo)
		userInfo.Username = reflect.ValueOf(sessionUser).String()
		err = userInfo.GetUserInfo()
		if err != nil {
			//need logging here instead of print
			fmt.Printf("Username not found, returning unauth, Get has %v\n", sessionStore)
			MyErrorResponse.Code = http.StatusUnauthorized
			MyErrorResponse.Error = "Session Expired"
			MyErrorResponse.HttpErrorResponder(w)
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
