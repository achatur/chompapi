package globalsessionkeeper

import (
	"github.com/astaxie/beego/session"
	_ "github.com/astaxie/beego/session/mysql"
	_ "fmt"
)

//Global Variable
var GlobalSessions *session.Manager