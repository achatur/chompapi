package session globalsessionkeeper

import (
	"github.com/astaxie/beego/session"
	_ "github.com/astaxie/beego/session/mysql"
	"fmt"
)

//Global Variable
var GlobalSessions *session.Manager