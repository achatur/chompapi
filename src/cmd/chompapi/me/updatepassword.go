package me
import (
	"encoding/json"
	"fmt"
	"net/http"
	"cmd/chompapi/db"
	"cmd/chompapi/globalsessionkeeper"
	"reflect"
	"unicode/utf8"
	"encoding/hex"
	"cmd/chompapi/crypto"
	"cmd/chompapi/messenger"
)

func UpdatePassword(w http.ResponseWriter, r *http.Request) {
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

		input := new(db.UserInfo)
		// dbUserInfo := new(db.UserInfo)
		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(&input); err != nil {
			fmt.Printf("something %v", err)
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Desc= "Malformed JSON: " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
		}

		fmt.Printf("Json Input = %+v\n", input)
		fmt.Printf("pass = %v\n", input.Password)

		if isValidInputPassword(input, &myErrorResponse) == false {
			fmt.Println("Something not valid")
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		input.Username = username

		if err := input.GetUserInfo(); err != nil {
			fmt.Printf("Could not find user")
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Desc= "User Not Found " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		input.PasswordHash = hex.EncodeToString(crypto.GeneratePassword(input.Username, []byte(input.Password)))
		fmt.Printf("Hash = %s\n", input.PasswordHash)

		if err := input.UpdatePassword(false); err != nil {
			fmt.Println("Error! = %v\n", err)
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Desc= "Could not Update Password: " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		// Send email
		fmt.Println("Sending Email...")
		body := fmt.Sprintf("Your password was recently changed.\n\nRegards,\n\nThe Chomp Team")
		context := new(messenger.SmtpTemplateData)
	    context.From = "The Chomp Team"
	    context.To = input.Email
	    context.Subject = "Password Changed"
	    context.Body = body

	    err := context.SendGmail()
	    if err != nil {
	    	fmt.Printf("Something ewnt wrong %v\n", err)
	    	myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Desc= "Could not send mail" + err.Error()
			myErrorResponse.HttpErrorResponder(w)
	    }

	    fmt.Printf("Mail sent")
		w.WriteHeader(http.StatusNoContent)
		return
		
	default:

		myErrorResponse.Code = http.StatusMethodNotAllowed
		myErrorResponse.Desc= "Invalid Method"
		myErrorResponse.HttpErrorResponder(w)
		return
	}

}

func isValidInputPassword(userInfo *db.UserInfo, errorResponse *globalsessionkeeper.ErrorResponse) bool {
	if isValidString(userInfo.Password) == false {
		fmt.Println("not valid Password = ", userInfo.Password)
		errorResponse.Error = "Invalid Password " + userInfo.Password
		return false
	} else if utf8.RuneCountInString(userInfo.Password) < 8 {
		errorResponse.Error = "Invalid Pass. Password must be at least 8 characters"
		return false
	}
	return true
}

func isValidString(s string) bool {
	fmt.Println("inside isValidString func")
	if s == "" {
		return false
	} else {
		return true
	}
}


