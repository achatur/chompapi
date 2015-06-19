package globalsessionkeeper

import (
	"github.com/astaxie/beego/session"
	_ "github.com/astaxie/beego/session/mysql"
	"encoding/json"
	"fmt"
	"net/http"
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