package main

import (
	//"crypto/sha1"
	"encoding/json"
	"fmt"
	//"hash"
	"net/http"
	//"text/template"
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
		validated := validatePassword(input.Username, []byte(input.Password), string(dbPassword))
		fmt.Printf("answer = %v", validatePassword(input.Username, []byte(input.Password), string(dbPassword)))
		if validated == true {
			sess, err := globalSessions.SessionStart(w, r)
			if err != nil {
				fmt.Printf("Error %v\n", err)
			}
			fmt.Printf("Session = %v\n", sess)
			//defer sess.SessionRelease(w)
			sess.Set("username", input.Username)
			username := sess.Get("username")
			fmt.Printf("Username = %v\n", username)

			//t, _ := template.ParseFiles("login.gtpl")
			//t.Execute(w, nil)
		} else {
			fmt.Printf("Login Failed")
			//return error
		}
	default:
		//	fmt.Fprintf(w, "Wrong Format")
		w.WriteHeader(http.StatusMethodNotAllowed)

	}
}
