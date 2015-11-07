package login

import (
	"encoding/json"
	"fmt"
	"net/http"
	"cmd/chompapi/db"
	"cmd/chompapi/crypto"
	_ "github.com/achatur/beego/session"
	"cmd/chompapi/globalsessionkeeper"
	"strconv"
	"time"
)

type LoginInput struct {
	Username string
	Password string
}

func DoLogin(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) (error) {

	switch r.Method {
	case "POST":
		var input LoginInput
		userInfo := new(db.UserInfo)
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			//need logging here instead of print
			fmt.Printf("something went wrong in login %v", err)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Malformed JSON: " + err.Error()}
		}

		fmt.Printf("input = %v\n", input)
		fmt.Printf("Number of active sessions: %v\n", globalsessionkeeper.GlobalSessions.GetActiveSession())
		userInfo.Username = input.Username
		userInfo.Email = input.Username
		err := userInfo.GetUserInfo(a.DB)
		if err != nil {
			//need logging here instead of print
			fmt.Println("Username not found..", input.Username)
			fmt.Println("Username not found..", input.Password)
			err := userInfo.GetUserInfoByEmailForLogin(a.DB)
			if err != nil {
				//need logging here instead of print
				fmt.Println("Email not found..", input.Username)
				fmt.Println("Email not found..", input.Password)
				err := userInfo.GetUserInfoByEmailForLogin(a.DB)
				return globalsessionkeeper.ErrorResponse{http.StatusUnauthorized, "Invalid Username"}
			}
		}
		fmt.Println("return from db = %v", userInfo)

		if (userInfo.IsPasswordTemp) {

			if userInfo.PasswordExpiry < int(time.Now().Unix()) {
				return globalsessionkeeper.ErrorResponse{http.StatusUnauthorized, "Temp Password Expired"}
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
			if err != nil {
				//need logging here instead of print
				fmt.Printf("Error, could not start session %v\n", err)
				return globalsessionkeeper.ErrorResponse{http.StatusUnauthorized, "Unauthorized"}
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
					return globalsessionkeeper.ErrorResponse{http.StatusUnauthorized, "Unauthorized"}
				}
				err = sessionStore.Set("userId", userInfo.UserID)
				if err != nil {
					//need logging here instead of print
					fmt.Printf("Error while writing to DB, %v\n", err)
					return globalsessionkeeper.ErrorResponse{http.StatusUnauthorized, "Unauthorized"}
				}
			} else {
				//need logging here instead of print
				fmt.Printf("Found Session! Session username = %v\n", sessionStore.Get("username"))
			}
		} else {
			fmt.Printf("Login Failed\n")
			return globalsessionkeeper.ErrorResponse{http.StatusUnauthorized, "Invalid Password"}
		}
		//Send back 204 no content (with cookie) + temp password header
		if userInfo.PasswordExpiry > 0 && userInfo.IsPasswordTemp == true {
			w.Header().Set("PasswordExpiry", strconv.Itoa(userInfo.PasswordExpiry))
			w.Header().Set("Location", "https://chompapi.com/me/update/up")
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	default:
		return globalsessionkeeper.ErrorResponse{http.StatusUnauthorized, "Unauthorized"}
	}
}
