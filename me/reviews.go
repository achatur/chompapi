package me
import (
	"encoding/json"
	"fmt"
	"net/http"
	"chompapi/db"
	"chompapi/globalsessionkeeper"
	"reflect"
)
type ReturnJson struct {
	Reviews []db.Review `json:"reviews"`
}
func Reviews(w http.ResponseWriter, r *http.Request) {
	var myErrorResponse globalsessionkeeper.ErrorResponse
	cookie := globalsessionkeeper.GetCookie(r)
	if cookie == "" {
			//need logging here instead of print
		fmt.Printf("Cookie = %v\n", cookie)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	sessionStore, err := globalsessionkeeper.GlobalSessions.GetSessionStore(cookie)
	if err != nil {
			//need logging here instead of print
			w.WriteHeader(http.StatusUnauthorized)
			return
	}
	sessionUser := sessionStore.Get("username")
	sessionUserID := sessionStore.Get("userId")
	fmt.Printf("SessionUser = %v\n", sessionUser)
	fmt.Printf("This SessionUserID = %v\n", sessionUserID)


	//reset time to time.now() + maxlifetime
	defer sessionStore.SessionRelease(w)

	//create variables
	userId := reflect.ValueOf(sessionUserID).Int()

	switch r.Method {

	case "GET":
		reviews := db.GetReviewsByUserID(int(userId))
		if reviews == nil {
			//something bad happened
			fmt.Printf("something went while retrieving data %v\n", err)
			fmt.Printf("Reviews list = %v\n", reviews)
			w.Header().Set("Content-Type", "application/json")
			emptyList := json.RawMessage(`{"reviews" : [] }`)
         	json.NewEncoder(w).Encode(&emptyList)
			return
		}
		fmt.Printf("Reviews list = %v\n", reviews)
		w.Header().Set("Content-Type", "application/json")
		returnJson :=  new(ReturnJson)
		returnJson.Reviews = reviews
		fmt.Printf("\n\nReview: reviews = %v\n", returnJson)
         json.NewEncoder(w).Encode(returnJson)
         if err != nil {
             fmt.Printf("something went while retrieving data %v\n", err)
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
            return
         }
         return

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

