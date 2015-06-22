package globalsessionkeeper

import (
	"github.com/astaxie/beego/session"
	_ "github.com/astaxie/beego/session/mysql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

//Global Variable
var GlobalSessions *session.Manager

type managerConfig struct {
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

type ErrorResponse struct {
	Code				int
	CustomMessage		string
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