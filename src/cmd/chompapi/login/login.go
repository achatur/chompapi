package login

import (
	"encoding/json"
	"fmt"
	"net/http"
	"cmd/chompapi/db"
	"cmd/chompapi/crypto"
	_ "github.com/astaxie/beego/session"
	"cmd/chompapi/globalsessionkeeper"
	"strconv"
	"time"
)

type LoginInput struct {
	Username string
	Password string
}

// var globalSessions *session.Manager

func DoLogin(w http.ResponseWriter, r *http.Request) {

	var myErrorResponse globalsessionkeeper.ErrorResponse

	switch r.Method {
	case "POST":
		var input LoginInput
		userInfo := new(db.UserInfo)
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			//need logging here instead of print
			fmt.Printf("something went wrong in login %v", err)
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "Malformed JSON: " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		fmt.Printf("input = %v\n", input)
		fmt.Printf("Number of active sessions: %v\n", globalsessionkeeper.GlobalSessions.GetActiveSession())
		userInfo.Username = input.Username
		err := userInfo.GetUserInfo()
		if err != nil {
			//need logging here instead of print
			fmt.Println("Username not found..", input.Username)
			fmt.Println("Username not found..", input.Password)
			myErrorResponse.Code = http.StatusUnauthorized
			myErrorResponse.Error = "Invalid Username"
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		fmt.Println("return from db = %v", userInfo)

		if (userInfo.IsPasswordTemp) {

			if userInfo.PasswordExpiry < int(time.Now().Unix()) {
				myErrorResponse.Code = http.StatusUnauthorized
				myErrorResponse.Error = "Temp Password Expired"
				myErrorResponse.HttpErrorResponder(w)
				return
			}

		}
		

		dbPassword := userInfo.PasswordHash

		validated := crypto.ValidatePassword(input.Username, []byte(input.Password), dbPassword)
		//need logging here instead of print or get rid of this statement in full once final
		fmt.Printf("answer = %v\n", validated)

		if validated == true {
			//create session using the request data which includes the cookie/sessionid
			fmt.Printf("about to start session\n")
			sessionStore, err := globalsessionkeeper.GlobalSessions.SessionStart(w, r)
			// sessionStore, err := GlobalSessions.SessionStart(w, r)
			if err != nil {
				//need logging here instead of print
				fmt.Printf("Error, could not start session %v\n", err)
				break
			}
			defer sessionStore.SessionRelease(w) //update db upon completion for request

			if sessionStore.Get("username") == nil {
				//need logging here instead of print
				fmt.Printf("Username not found, Saving Session, Get has %v\n", sessionStore)
				fmt.Printf("Username not found, Saving Session, Get has %v\n", sessionStore.Get("usernamestring"))
				err = sessionStore.Set("username", input.Username)
				if err != nil {
					//need logging here instead of print
					fmt.Printf("Error while writing to DB, %v\n", err)
				}
				err = sessionStore.Set("userId", userInfo.UserID)
				if err != nil {
					//need logging here instead of print
					fmt.Printf("Error while writing to DB, %v\n", err)
					break
				}
			} else {
				//need logging here instead of print
				fmt.Printf("Found Session! Session username = %v\n", sessionStore.Get("username"))
			}
		} else {
			fmt.Printf("Login Failed")
			w.WriteHeader(http.StatusUnauthorized)
			myErrorResponse.Code = http.StatusUnauthorized
			myErrorResponse.Error = "Invalid Password"
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		//Send back 204 no content (with cookie) + temp password header
		if userInfo.PasswordExpiry > 0 && userInfo.IsPasswordTemp == true {
			w.Header().Set("PasswordExpiry", strconv.Itoa(userInfo.PasswordExpiry))
			w.Header().Set("Location", "https://chompapi.com/me/update/up")
		}
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		// w.WriteHeader(http.StatusUnauthorized)
		myErrorResponse.Code = http.StatusUnauthorized
		myErrorResponse.HttpErrorResponder(w)
		return
	}
}
