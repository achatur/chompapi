package globalsessionkeeper

import (
	"github.com/astaxie/beego/session"
	_ "github.com/astaxie/beego/session/mysql"
	"fmt"
)

//Global Variable
var GlobalSessions *session.Manager

func main() {}

func init() {
	var err error
	GlobalSessions, err = session.NewManager("mysql", `{"enableSetCookie":true, "SessionOn":true, "cookieName":"chomp_sessionid","gclifetime":120,"ProviderConfig":"root@tcp(172.16.0.1:3306)/chomp"}`)
	if err != nil {
		fmt.Printf("Error")
	}
	GlobalSessions.SetSecure(true)
	go GlobalSessions.GC()
}
