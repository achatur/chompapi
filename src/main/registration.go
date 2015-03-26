package main

import (
    "fmt"
    "net/http"
    "encoding/json"
)

type RegisterInput struct {
    FName string
    Lname string
    Email string
    Username string
    Password string
    Dob string
    Gender string
}

func doRegister(w http.ResponseWriter, r * http.Request) {
   //fmt.Fprintf(w, "Hello Registerer") 

    switch r.Method {
    case "POST":
        var input RegisterInput
        decoder := json.NewDecoder(r.Body)
        if err := decoder.Decode(&input); err != nil {fmt.Printf("something %v", err)}
        fmt.Printf("%+v", input)

    default:
    //    fmt.Fprintf(w, "Wrong Format")
        w.WriteHeader(http.StatusMethodNotAllowed)

    
    }
}
