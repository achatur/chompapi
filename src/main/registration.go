package main

import (
	//"database/sql"
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
	//fmt.Fprintf(w, "Hello Registerer")

	switch r.Method {
	case "POST":
		//var input RegisterInput
		input := NewUser()
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			fmt.Printf("something %v", err)
		}
		m := map[string]string{}
		m["username"] = "amir_test"
		m["email"] = "amir@chomp.com"
		m["password_hash"] = "password"
		m["phone"] = "1230404049"
		m["dob"] = "null"
		m["gender"] = "m"
		fmt.Printf("%+v", input)
		fmt.Printf("%+v", m)
		decodedHexString, err_ := hex.DecodeString(input.Hash)
		if err_ != nil {
			fmt.Println("Error! = %v\n", err_)
		}
		//input.Hash = string(generatePassword(input.Username, []byte(decodedHexString)))
		//input.Hash = hex.Dump(generatePassword(input.Username, []byte(decodedHexString)))
		if isValidInput(input.Email) == false {
			fmt.Println("made it here, value = ", input.Email)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		input.Hash = hex.EncodeToString(generatePassword(input.Username, []byte(decodedHexString)))
		fmt.Printf("Hash = %s\n", input.Hash)
		err := input.SetUserInfo()
		if err != nil {
			fmt.Println("Error! = %v\n", err)
		}

	default:
		//	fmt.Fprintf(w, "Wrong Format")
		w.WriteHeader(http.StatusMethodNotAllowed)

	}
}
func isValidInput(s string) bool {
	fmt.Println("inside isValidInput func")
	if s == "" {
		return false
	} else {
		return true
	}
	//if ns.Valid {
	//	return true
	//} else if ns.String == "" {
	//	fmt.Println("ns.String = ", ns.String)
	//	return false
	//} else {
	//	fmt.Println("true ns.String = ", ns.String)
	//	return true
	//}
}

func NewUser() *RegisterInput {
	//return &RegisterInput{Fname: "null", Lname: "null", Email: "null", Username: "null", Password: "null", Dob: "null", Gender: "null", Phone: "null"}
	return &RegisterInput{}
}
