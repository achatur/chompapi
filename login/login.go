package login

import (
	"encoding/json"
	"fmt"
	"net/http"
	"chompapi/db"
	"chompapi/crypto"
	"github.com/astaxie/beego/session"
	"chompapi/globalsessionkeeper"
)

type LoginInput struct {
	Username string
	Password string
}

// type UserInfo struct {
// 	ChompUserID   int
// 	ChompUsername string
// 	Email         string
// 	PhoneNumber   string
// 	PasswordHash  string
// 	DOB           string
// 	Gender        string
// }

var globalSessions *session.Manager

func DoLogin(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":
		var input LoginInput
		userInfo := new(db.UserInfo)
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			//need logging here instead of print
			fmt.Printf("something went wrong in login %v", err)
			break
		}

		fmt.Printf("input = %v\n", input)
		fmt.Printf("Number of active sessions: %v\n", globalsessionkeeper.GlobalSessions.GetActiveSession())
		userInfo.ChompUsername = input.Username
		err := userInfo.GetUserInfo(input.Username)
		if err != nil {
			//need logging here instead of print
			fmt.Println("Username not found..", input.Username)
			fmt.Println("Username not found..", input.Password)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Println("return from db = %v", userInfo)

		// dbPassword := userInfo["password_hash"]
		dbPassword := userInfo.PasswordHash

		validated := crypto.ValidatePassword(input.Username, []byte(input.Password), dbPassword)
		//need logging here instead of print or get rid of this statement in full once final
		fmt.Printf("answer = %v\n", validated)

		if validated == true {
			//create session using the request data which includes the cookie/sessionid
			// fmt.Printf("Manager Config = %v", globalsessionkeeper.GlobalSessions.config)
			fmt.Printf("about to start session\n")
			sessionStore, err := globalsessionkeeper.GlobalSessions.SessionStart(w, r)
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
				err = sessionStore.Set("userID", userInfo.ChompUserID)
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
			return
		}
		//Send back 204 no content (with cookie)
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	return
}
