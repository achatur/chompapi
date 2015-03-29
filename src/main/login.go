package main

import (
	//"crypto/sha1"
	"encoding/json"
	"fmt"
	//"hash"
	"net/http"
)

type LoginInput struct {
	Username string
	Password string
}

func doLogin(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hello Registerer")

	switch r.Method {
	case "POST":
		var input LoginInput
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			fmt.Printf("something %v", err)
		}
		fmt.Printf("%+v", input)
		dbPassword := generatePassword(input.Username, []byte(input.Password))
		fmt.Printf("dbPass= %+x\n", dbPassword)
		fmt.Printf("answer = %v", validatePassword(input.Username, []byte(input.Password), string(dbPassword)))

	default:
		//	fmt.Fprintf(w, "Wrong Format")
		w.WriteHeader(http.StatusMethodNotAllowed)

	}
}
