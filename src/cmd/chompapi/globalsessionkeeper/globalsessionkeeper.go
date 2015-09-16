package globalsessionkeeper

import (
	"github.com/achatur/beego/session"
	_ "github.com/achatur/beego/session/mysql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"errors"
	"database/sql"
)

type AppContext struct {
    DB        *sql.DB
    // store     *sessions.CookieStore
    SessionStore session.SessionStore
    Manager 	*session.Manager
    // templates map[string]*template.Template
    // decoder   *schema.Decoder
    // store     *redistore.RediStore
    // mandrill  *gochimp.MandrillAPI
    // twitter   *anaconda.TwitterApi
    // log       *log.Logger
    // config      *ChompConfig // app-wide configuration: hostname, ports, etc.
    // MyErrorResponse globalsessionkeeper.ErrorResponse
}

type Config struct {
	Authorized 		[]Authorized 	`json:"authorized"`
	Cert 	 		PrivateKey 		`json:"privateKey"`
	DbConfig 		DbConfig 		`json:"dbConfig"`
	ManagerConfig	ManagerConfig	`json:"managerConfig"`
}
type Authorized struct {
	User 		string 		`json:"user"`
	Pass 		string 		`json:"pass"`
}

type PrivateKey struct {
	Cert		string `json:"cert"`
	Key 		string `json:"key"`
}

type DbConfig struct {
	Type 		string `json:"type"`
	User 		string `json:"user"`
	Pass	 	string `json:"pass"` 
	Host 		string `json:"host"`
	Port 		string `json:"port"`
	Db 			string `json:"db"`
}

type ManagerConfig struct {
	CookieName      string `json:"cookieName"`
	EnableSetCookie bool   `json:"enableSetCookie,omitempty"`
	Gclifetime      int64  `json:"gclifetime"`
	Maxlifetime     int64  `json:"maxLifetime"`
	Secure          bool   `json:"secure"`
	CookieLifeTime  int    `json:"cookieLifeTime"`
	ProviderConfig  string `json:"providerConfig"`
	Domain          string `json:"domain"`
	SessionIdLength int64  `json:"sessionIdLength"`
}

//Global Variable
var GlobalSessions *session.Manager
var ChompConfig Config

type ErrorResponse struct {
	Code				int `json:"code"`
	Desc 				string `json:"error"`
}

func (h ErrorResponse) Error() string {
    return fmt.Sprintf("HTTP %d: %s", h.Code, h.Error)
}

func (errorResponse ErrorResponse) HttpErrorResponder(w http.ResponseWriter) {

	fmt.Print("Going out as: %v\n", errorResponse)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorResponse.Code)
	json.NewEncoder(w).Encode(errorResponse)
}

func GetCookie(r *http.Request) string {

	fmt.Println("Full header = %v", r.Header)
	cookie, err := r.Cookie("chomp_sessionid")
	if err != nil {
		fmt.Println("Error..cookie = %v, err:%v, cookie1:%v err1:%v",cookie, err)
		return ""
	}
	fmt.Println("Cookie = %v", cookie)

	if cookiestr := r.Header.Get("Cookie"); cookiestr == "" {
		return ""
	} else {
		parts := strings.Split(strings.TrimSpace(cookiestr), ";")
		for k, v := range parts {
			nameval := strings.Split(v, "=")
			if k == 0 && nameval[0] != "chomp_sessionid" {
				return ""
			} else {
				fmt.Printf("Returning cookie %v\n", nameval[1])
				return nameval[1]
			}
		}
	}
	return ""
}

func ExpireCookie(r *http.Request, w http.ResponseWriter) error {

	fmt.Println("Full header = %v", r.Header)
	cookie, err := r.Cookie("chomp_sessionid")
	if err != nil {
		fmt.Println("Error..cookie = %v, err:%v, cookie1:%v err1:%v",cookie, err)
		return err
	}
	fmt.Println("Cookie = %v", cookie)

	if cookiestr := r.Header.Get("Cookie"); cookiestr == "" {
		return errors.New("Cookie not found")
	} else {
		expiration := time.Now()
		cookie.Expires = expiration
		http.SetCookie(w, cookie)
	}
	return nil
}