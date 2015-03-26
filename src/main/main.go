package main

import (
    "fmt"
    "html"
    "log"
    "net/http"

)

func main() {
    http.HandleFunc("/register", doRegister)
    http.HandleFunc("/login", doLogin)

    log.Fatal(http.ListenAndServe(":8080", nil))
}

func Index(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}


func doLogin(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "welcome, %q", html.EscapeString(r.URL.Path))
}
