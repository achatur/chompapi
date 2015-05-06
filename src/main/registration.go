package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
)

type RegisterInput struct {
	Username string
	Email    string
	Password string
	Dob      string
	Gender   string
	Fname    string
	Lname    string
	Phone    string
	Hash     string
}

func doRegister(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":
		input := NewUser()
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			fmt.Printf("something %v", err)
		}
		fmt.Printf("%+v", input)
		if isValidInput(input) == false {
			fmt.Println("Something not valid")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		input.Hash = hex.EncodeToString(generatePassword(input.Username, []byte(input.Password)))
		fmt.Printf("Hash = %s\n", input.Hash)
		err := input.SetUserInfo()
		if err != nil {
			fmt.Println("Error! = %v\n", err)
			if err.Contains("Error 1062") {
				w.WriteHeader(http.StatusConflict)
			}
			w.WriteHeader(http.StatusInternalServerError)
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
func isValidInput(userInfo *RegisterInput) bool {
	if isValidString(userInfo.Email) == false {
		fmt.Println("not valid email = ", userInfo.Email)
		return false
	}
	if isValidString(userInfo.Username) == false {
		fmt.Println("not valid username", userInfo.Email)
		return false
	}
	if isValidString(userInfo.Password) == false {
		fmt.Println("not valid password", userInfo.Email)
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

func NewUser() *RegisterInput {
	return &RegisterInput{}
}
