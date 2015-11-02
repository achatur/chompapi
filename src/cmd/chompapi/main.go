package main

import (
	"fmt"
	// "html"
	"log"
	"net/http"
	"os"
	"strings"
	"cmd/chompapi/login"
	"cmd/chompapi/register"
	"cmd/chompapi/globalsessionkeeper"
	"github.com/achatur/beego/session"
	"cmd/chompapi/me"
	"cmd/chompapi/review"
	"database/sql"
	"github.com/gorilla/mux"
	"cmd/chompapi/crypto"
	"encoding/base64"
	"io/ioutil"
	"encoding/json"
	"cmd/chompapi/db"
	"reflect"
	"errors"
)

type handler func(w http.ResponseWriter, r *http.Request)
var MyDb *sql.DB
// var context *globalsessionkeeper.AppContext

func init() {

	var err error
	GetConfig()
	sessionConfig, _ := json.Marshal(globalsessionkeeper.ChompConfig.ManagerConfig)
	dbConfig := globalsessionkeeper.ChompConfig.DbConfig
	fmt.Printf("\n\n\nIn init, new manager\n")
	fmt.Printf("In init, new manager\n")
	fmt.Printf("In init, new manager\n\n\n\n")
	globalsessionkeeper.GlobalSessions, err = session.NewManager("mysql", string(sessionConfig))

	if err != nil {
		fmt.Printf("Coud not start session..Error: %v\n", err.Error())
		os.Exit(-1)

	}
	err = errors.New("")
	fmt.Printf("Opening DB connection\n")
	// Connection string looks as the following
	//MyDb, err = sql.Open("service", "user@tcp(ip:port)/database")
	connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbConfig.User, dbConfig.Pass, dbConfig.Host,dbConfig.Port, dbConfig.Db)
	fmt.Printf("ConnString = %s\n", connString)
	MyDb, err = sql.Open("mysql", connString) 
	if err != nil {
		// return err
		fmt.Printf("Error = %v\n", err)
		panic(fmt.Sprintf("%v", err))
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

func (ah AppHandler) SessionAuth(pass handler) handler {

	return func(w http.ResponseWriter, r *http.Request) {
		cookie := globalsessionkeeper.GetCookie(r)
		if cookie == "" {
			//need logging here instead of print
			fmt.Println("Session Auth Cookie = %v", cookie)
			query := mux.Vars(r)
			fmt.Printf("Query here.. %v\n", query)
			if query["token"] != "" {
				fmt.Printf("Error not nil, updating error instacode %v\n", query["token"])
				cookie = query["token"]
			} else {
				HttpErrorResponder(w, globalsessionkeeper.ErrorResponse{http.StatusUnauthorized, "No Cookie Present"})
				return
			}
		}
	
		sessionStore, err := globalsessionkeeper.GlobalSessions.GetSessionStore(cookie)
		if err != nil {
			//need logging here instead of print
			HttpErrorResponder(w, globalsessionkeeper.ErrorResponse{http.StatusUnauthorized, "Session Expired"})
			return
		}
		defer sessionStore.SessionRelease(w)
		ah.appContext.SessionStore = sessionStore
		sessionUser := sessionStore.Get("username")
		fmt.Printf("Session Auth SessionUser = %v\n", sessionUser)
		if sessionUser == nil {
			//need logging here instead of print
			fmt.Printf("Username not found, returning unauth, Get has %v\n", sessionStore)
			HttpErrorResponder(w, globalsessionkeeper.ErrorResponse{http.StatusUnauthorized, "Session Expired"})
			return
		}

		fmt.Printf("Session Auth Getting user info for user %v\n", sessionUser)
		userInfo := new(db.UserInfo)
		userInfo.Username = reflect.ValueOf(sessionUser).String()
		err = userInfo.GetUserInfo(MyDb)
		if err != nil {
			//need logging here instead of print
			fmt.Printf("Session Auth Username not found, returning unauth, Get has %v\n", sessionStore)
			HttpErrorResponder(w, globalsessionkeeper.ErrorResponse{http.StatusUnauthorized, "Session Expired"})
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


type AppHandler struct { 
	appContext *globalsessionkeeper.AppContext
	h func(*globalsessionkeeper.AppContext, http.ResponseWriter, *http.Request) (error)
}

func HttpErrorResponder(w http.ResponseWriter, errorResponse globalsessionkeeper.ErrorResponse) {

	fmt.Print("Going out as: %v\n", errorResponse)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorResponse.Code)
	json.NewEncoder(w).Encode(errorResponse)
}

func (ah AppHandler) ServerHttp(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("AH Context = %v\n", ah.appContext)
	err := ah.h(ah.appContext, w, r)
	if err != nil {
		// log.Printf("HTTP %d: %q", status, err)
		status := err.(globalsessionkeeper.ErrorResponse).Code
		switch status {
		case http.StatusNotFound:
			fmt.Printf("Error: Page not found\n")
			HttpErrorResponder(w, err.(globalsessionkeeper.ErrorResponse))
		case http.StatusInternalServerError:
			fmt.Printf("Error: %v\n", http.StatusInternalServerError)
			HttpErrorResponder(w, err.(globalsessionkeeper.ErrorResponse))
		default:
			fmt.Printf("Error: %v\n", err)
			HttpErrorResponder(w, err.(globalsessionkeeper.ErrorResponse))
		}
	}
}

func main() {

	defer MyDb.Close()

	router := mux.NewRouter().StrictSlash(true)
	context := &globalsessionkeeper.AppContext{DB: MyDb}
	fmt.Printf("Context = %v\n", context)


	router.HandleFunc("/login", AppHandler{context, login.DoLogin}.ServerHttp)
	router.HandleFunc("/register", AppHandler{context, register.DoRegister}.ServerHttp)

	router.HandleFunc("/admin/fp", BasicAuth(AppHandler{context, register.ForgotPassword}.ServerHttp))
	router.HandleFunc("/admin/fu", BasicAuth(AppHandler{context, register.ForgotUsername}.ServerHttp))
	router.HandleFunc("/admin/jwt", BasicAuth(AppHandler{context, crypto.GetJwt}.ServerHttp))

	router.HandleFunc("/me", AppHandler{appContext: context, h: me.GetMe}.SessionAuth(AppHandler{appContext: context, h: me.GetMe}.ServerHttp))

	//this is how you write a query parameter capture uri
	router.Queries("token", "{token}").HandlerFunc(AppHandler{appContext: context, h: me.Instagram}.SessionAuth(AppHandler{context, me.Instagram}.ServerHttp))
	router.Queries("token", "cd80fd58ebf00b8bf7ff2a4020fa2a9b", "code", "{code:.*").HandlerFunc(AppHandler{appContext: context, h: me.Instagram}.SessionAuth(AppHandler{context, me.Instagram}.ServerHttp))
	router.Queries("token", "{token}", "error", "{error").HandlerFunc(AppHandler{appContext: context, h: me.Instagram}.SessionAuth(AppHandler{context, me.Instagram}.ServerHttp))
	// router.Queries("code", "{code}").HandlerFunc(AppHandler{appContext: context, h: me.Instagram}.SessionAuth(AppHandler{context, me.Instagram}.ServerHttp))
	router.Queries("error", "{error}").HandlerFunc(AppHandler{appContext: context, h: me.Instagram}.SessionAuth(AppHandler{context, me.Instagram}.ServerHttp))

	router.HandleFunc("/me/logout", AppHandler{appContext: context, h: me.Logout}.SessionAuth(AppHandler{appContext: context, h: me.Logout}.ServerHttp))
	router.HandleFunc("/me/logout/all", AppHandler{appContext: context, h: me.LogoutAll}.SessionAuth(AppHandler{appContext: context, h: me.LogoutAll}.ServerHttp))

	router.HandleFunc("/me/photos", AppHandler{appContext: context, h: me.PostPhotoId}.SessionAuth(AppHandler{appContext: context, h: me.PostPhotoId}.ServerHttp))
	router.HandleFunc("/me/photos/{photoID}", AppHandler{appContext: context, h: me.PostPhotoId}.SessionAuth(AppHandler{appContext: context, h: me.PostPhotoId}.ServerHttp))

	router.HandleFunc("/me/reviews", AppHandler{appContext: context, h: me.Reviews}.SessionAuth(AppHandler{appContext: context, h: me.Reviews}.ServerHttp))

	router.HandleFunc("/me/update/up", AppHandler{appContext: context, h: me.UpdatePassword}.SessionAuth(AppHandler{appContext: context, h: me.UpdatePassword}.ServerHttp))
	router.HandleFunc("/me/update/d/{userID}", AppHandler{appContext: context, h: me.DeleteMe}.SessionAuth(AppHandler{appContext: context, h: me.DeleteMe}.ServerHttp))
	router.HandleFunc("/me/update/instaClick", AppHandler{appContext: context, h: me.InstagramLinkClick}.SessionAuth(AppHandler{appContext: context, h: me.InstagramLinkClick}.ServerHttp))

	router.HandleFunc("/me/update/da/{userID}", AppHandler{appContext: context, h: me.DeactivateMe}.SessionAuth(AppHandler{appContext: context, h: me.DeactivateMe}.ServerHttp))
	router.HandleFunc("/me/update/astu", AppHandler{appContext: context, h: me.UpdateAccountSetupTimestamp}.SessionAuth(AppHandler{appContext: context, h: me.UpdateAccountSetupTimestamp}.ServerHttp))

	router.HandleFunc("/reviews", AppHandler{appContext: context, h: review.Reviews}.SessionAuth(AppHandler{appContext: context, h: review.Reviews}.ServerHttp))
	router.HandleFunc("/reviews/{reviewID}", AppHandler{appContext: context, h: review.Reviews}.SessionAuth(AppHandler{appContext: context, h: review.Reviews}.ServerHttp))
	
	router.HandleFunc("/insta/crawl", AppHandler{appContext: context, h: review.Crawl}.SessionAuth(AppHandler{appContext: context, h: review.Crawl}.ServerHttp))


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
