package me
import (
	"encoding/json"
	"fmt"
	"net/http"
	"cmd/chompapi/db"
	"cmd/chompapi/globalsessionkeeper"
	"reflect"
)
type ReturnJson struct {
	Reviews []db.Review `json:"reviews"`
}
func Reviews(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {
	cookie := globalsessionkeeper.GetCookie(r)
	if cookie == "" {
			//need logging here instead of print
		return globalsessionkeeper.ErrorResponse{http.StatusUnauthorized, "No Cookie Present"}
	}
	sessionStore, err := globalsessionkeeper.GlobalSessions.GetSessionStore(cookie)
	if err != nil {
			//need logging here instead of print
		return globalsessionkeeper.ErrorResponse{http.StatusUnauthorized, "Expired Cookie Provided"}
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
		reviews, err := db.GetReviewsByUserID(int(userId), a.DB)
		if err != nil {
			fmt.Printf("something went while retrieving data %v\n", err)
			fmt.Printf("Reviews list = %v\n", reviews)
			w.Header().Set("Content-Type", "application/json")
			returnJson := reviews
         	json.NewEncoder(w).Encode(&returnJson)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, err.Error()}
		}
		if reviews == nil {
			//something bad happened
			fmt.Printf("something went while retrieving data %v\n", err)
			fmt.Printf("Reviews list = %v\n", reviews)
			w.Header().Set("Content-Type", "application/json")
			emptyList := json.RawMessage(`{"reviews" : [] }`)
         	json.NewEncoder(w).Encode(&emptyList)
			return nil
		}
		fmt.Printf("Reviews list = %v\n", reviews)
		w.Header().Set("Content-Type", "application/json")
		returnJson :=  new(ReturnJson)
		returnJson.Reviews = reviews
		fmt.Printf("\n\nReview: reviews = %v\n", returnJson)
         json.NewEncoder(w).Encode(returnJson)
         if err != nil {
             fmt.Printf("something went while retrieving data %v\n", err)
             return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
         }
         return nil

	default:
		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Method Not Allowed"}
	}
}

