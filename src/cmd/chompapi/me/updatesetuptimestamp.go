package me
import (
	"fmt"
	"net/http"
	"cmd/chompapi/db"
	"cmd/chompapi/globalsessionkeeper"
	"reflect"
)

func UpdateAccountSetupTimestamp(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {
	sessionUser := a.SessionStore.Get("username")
	sessionUserID := a.SessionStore.Get("userId")
	fmt.Printf("SessionUser = %v\n", sessionUser)
	fmt.Printf("This SessionId = %v\n", sessionUserID)

	//create variables
	username := reflect.ValueOf(sessionUser).String()
	switch r.Method {
	case "PUT":

		// input := new(db.UserInfo)
		dbUserInfo := new(db.UserInfo)
		dbUserInfo.Username = username
		err := dbUserInfo.GetUserInfo(a.DB)
		if err != nil {
			fmt.Printf("Failed to get userinfo, err = %v\n", err)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, err.Error()}
		}
		dbUserInfo.Username = username

		fmt.Printf("Json Input = %+v\n", dbUserInfo)
		fmt.Printf("pass = %v\n", dbUserInfo.Password)

		err = dbUserInfo.UpdateAccountSetupTimestamp(a.DB)

		if err != nil {
			fmt.Println("Something not valid")
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, err.Error()}
		}

		return nil
		
	default:

		return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Method Not Allowed"}
	}

}
