package register

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"chompapi/db"
	"chompapi/crypto"
	"time"
	"chompapi/globalsessionkeeper"
)

func DoRegister(w http.ResponseWriter, r *http.Request) {
	var myErrorResponse globalsessionkeeper.ErrorResponse

	switch r.Method {
	case "POST":
		input := newUser()
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			fmt.Printf("something %v", err)
		}
		fmt.Printf("Json Input = %+v\n", input)
		fmt.Println("int = %v", input.Dob)

		if isValidInput(input, &myErrorResponse) == false {
			fmt.Println("Something not valid")
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.HttpErrorResponder(w)

			return
		}

		input.Hash = hex.EncodeToString(crypto.GeneratePassword(input.Username, []byte(input.Password)))
		fmt.Printf("Hash = %s\n", input.Hash)

		err := input.SetUserInfo()
		if err != nil {
			fmt.Println("Error! = %v\n", err)
			if strings.Contains(err.Error(), "Error 1062") {
				myErrorResponse.Code = http.StatusConflict
				myErrorResponse.CustomMessage = "Duplicate Not Allowed::ErrorMessage::" + err.Error()
				myErrorResponse.HttpErrorResponder(w)
				return
			}

			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.CustomMessage = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return

	default:

		myErrorResponse.Code = http.StatusMethodNotAllowed
		myErrorResponse.CustomMessage = "Invalid Method"
		myErrorResponse.HttpErrorResponder(w)
		return
	}
}

func isValidInput(userInfo *db.RegisterInput, errorResponse *globalsessionkeeper.ErrorResponse) bool {
	if isValidString(userInfo.Email) == false {
		fmt.Println("not valid email = ", userInfo.Email)
		errorResponse.CustomMessage = "Invalid Email " + userInfo.Email
		return false
	}
	if isValidString(userInfo.Username) == false {
		fmt.Println("not valid username", userInfo.Username)
		errorResponse.CustomMessage = "Invalid Username " + userInfo.Username
		return false
	}
	if isValidString(userInfo.Password) == false {
		fmt.Println("not valid password", userInfo.Password)
		errorResponse.CustomMessage = "Invalid Password " + userInfo.Password
		return false
	}
	if userInfo.Dob == 0 || age(time.Unix(int64(userInfo.Dob), 0)) < 18 {
		errorResponse.CustomMessage = "Invalid Age " + string(age(time.Unix(int64(userInfo.Dob), 0)))
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

func newUser() *db.RegisterInput {
	return &db.RegisterInput{}
}

func age(birthday time.Time) int {
	fmt.Println("made it here")
	now := time.Now()
	years := now.Year() - birthday.Year()
	if now.YearDay() < birthday.YearDay(){
		years--
	}
	fmt.Println("Age = %v", years)
	return years
}
