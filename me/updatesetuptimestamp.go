package me
import (
	"fmt"
	"net/http"
	"chompapi/db"
	"chompapi/globalsessionkeeper"
	"reflect"
)

func UpdateAccountSetupTimestamp(w http.ResponseWriter, r *http.Request) {
	var myErrorResponse globalsessionkeeper.ErrorResponse
	cookie := globalsessionkeeper.GetCookie(r)
	if cookie == "" {
			//need logging here instead of print
		fmt.Printf("Cookie = %v\n", cookie)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	sessionStore, err := globalsessionkeeper.GlobalSessions.GetSessionStore(cookie)

	if err != nil {
			//need logging here instead of print
			w.WriteHeader(http.StatusUnauthorized)
			return
	}

	sessionUser := sessionStore.Get("username")
	sessionUserID := sessionStore.Get("userId")
	fmt.Printf("SessionUser = %v\n", sessionUser)
	fmt.Printf("This SessionId = %v\n", sessionUserID)


	defer sessionStore.SessionRelease(w)
	//create variables
	username := reflect.ValueOf(sessionUser).String()
	switch r.Method {
	case "PUT":

		// input := new(db.UserInfo)
		dbUserInfo := new(db.UserInfo)
		dbUserInfo.Username = username
		err = dbUserInfo.GetUserInfo()
		if err != nil {
			fmt.Printf("Failed to get userinfo, err = %v\n", err)
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		dbUserInfo.Username = username

		fmt.Printf("Json Input = %+v\n", dbUserInfo)
		fmt.Printf("pass = %v\n", dbUserInfo.Password)

		err = dbUserInfo.UpdateAccountSetupTimestamp()

		if err != nil {
			fmt.Println("Something not valid")
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
		
	default:

		myErrorResponse.Code = http.StatusMethodNotAllowed
		myErrorResponse.Error = "Invalid Method"
		myErrorResponse.HttpErrorResponder(w)
		return
	}

}
