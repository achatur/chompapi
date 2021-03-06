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

func UpdatePassword(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {
	var myErrorResponse globalsessionkeeper.ErrorResponse
	sessionUser := a.SessionStore.Get("username")
	sessionUserID := a.SessionStore.Get("userId")
	fmt.Printf("SessionUser = %v\n", sessionUser)
	fmt.Printf("This SessionId = %v\n", sessionUserID)

	//create variables
	username := reflect.ValueOf(sessionUser).String()
	switch r.Method {
	case "PUT":

		input := new(db.UserInfo)
		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(&input); err != nil {
			fmt.Printf("something %v", err)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Malformed JSON: " + err.Error()}
		}

		fmt.Printf("Json Input = %+v\n", input)
		fmt.Printf("pass = %v\n", input.Password)

		if isValidInputPassword(input, &myErrorResponse) == false {
			fmt.Println("Something not valid")
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, myErrorResponse.Desc}
		}

		input.Username = username

		if err := input.GetUserInfo(a.DB); err != nil {
			fmt.Printf("Could not find user")
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "User Not Found " + err.Error()}
		}

		input.PasswordHash = hex.EncodeToString(crypto.GeneratePassword(input.Username, []byte(input.Password)))
		fmt.Printf("Hash = %s\n", input.PasswordHash)

		if err := input.UpdatePassword(false, a.DB); err != nil {
			fmt.Println("Error! = %v\n", err)
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "Could not Update Password: " + err.Error()}
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
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "Could not send mail" + err.Error()}
	    }

	    fmt.Printf("Mail sent")
		w.WriteHeader(http.StatusNoContent)
		return nil
		
	default:

		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Method Not Allowed"}
	}

}

func isValidInputPassword(userInfo *db.UserInfo, errorResponse *globalsessionkeeper.ErrorResponse) bool {
	if isValidString(userInfo.Password) == false {
		fmt.Println("not valid Password = ", userInfo.Password)
		errorResponse.Desc = "Invalid Password " + userInfo.Password
		return false
	} else if utf8.RuneCountInString(userInfo.Password) < 8 {
		errorResponse.Desc = "Invalid Pass. Password must be at least 8 characters"
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


