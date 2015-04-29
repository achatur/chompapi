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
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
func isValidInput(userInfo *RegisterInput) {
	if isValidString(input.Email) == false {
			fmt.Println("not valid email = ", input.Email)
			return false
	}
	if isValidString(input.Username) == false {
			fmt.Println("not valid username", input.Email)
			return false
	}
	if isValidString(input.Password) == false {
			fmt.Println("not valid password", input.Email)
			return false
	}
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
