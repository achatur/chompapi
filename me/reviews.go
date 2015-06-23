package me
import (
	"encoding/json"
	"fmt"
	"net/http"
	"chompapi/db"
	"chompapi/globalsessionkeeper"
	"reflect"
)
func Reviews(w http.ResponseWriter, r *http.Request) {
	var myErrorResponse globalsessionkeeper.ErrorResponse
	cookie := globalsessionkeeper.GetCookie(r)
	if cookie == "" {
			//need logging here instead of print
		fmt.Println("Cookie = %v", cookie)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	sessionStore, err := globalsessionkeeper.GlobalSessions.GetSessionStore(cookie)
	if err != nil {
			//need logging here instead of print
			w.WriteHeader(http.StatusUnauthorized)
			return
	}
	//input.Username = sessionStore.Get("username")
	sessionUser := sessionStore.Get("username")
	sessionUserID := sessionStore.Get("userID")
	fmt.Println("SessionUser = %v", sessionUser)

	if sessionUser == nil {
			//need logging here instead of print
			fmt.Printf("Username not found, returning unauth, Get has %v\n", sessionStore)
			w.WriteHeader(http.StatusUnauthorized)
			return
	} else {
		//reset time to time.now() + maxlifetime
		defer sessionStore.SessionRelease(w)

		//create variables
		userID 	 	 := reflect.ValueOf(sessionUserID).Int()

		switch r.Method {

		case "GET":
			reviews := db.GetReviewsByUserID(int(userID))
			if reviews == nil {
				//something bad happened
				fmt.Printf("something went while retrieving data %v\n", err)
				fmt.Printf("Reviews list = %v", reviews)
				w.Header().Set("Content-Type", "application/json")
            	json.NewEncoder(w).Encode("[]")
				return
			}
			fmt.Printf("Reviews list = %v", reviews)
			w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(reviews)
            if err != nil {
                fmt.Printf("something went while retrieving data %v\n", err)
				myErrorResponse.Code = http.StatusInternalServerError
				myErrorResponse.CustomMessage = err.Error()
				myErrorResponse.HttpErrorResponder(w)
                return
            }
            return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}

}

