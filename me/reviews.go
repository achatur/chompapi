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
	cookie := getCookie(r)
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
	fmt.Println("SessionUser = %v", sessionUser)

	if sessionUser == nil {
			//need logging here instead of print
			fmt.Printf("Username not found, returning unauth, Get has %v\n", sessionStore)
			w.WriteHeader(http.StatusUnauthorized)
			return
	} else {
		username := reflect.ValueOf(sessionUser).String()
		switch r.Method {

		case "POST":
			var sentRestaurant db.Restaurants
			var dbRestaurant db.Restaurants
			var review db.Reviews
			review.Username = username
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&sentRestaurant); err != nil {
				//need logging here instead of print
				fmt.Printf("something went wrong in login %v", err)
				myErrorResponse.Code = http.StatusInternalServerError
				myErrorResponse.CustomMessage = "Malformed Json" + err.Error()
				myErrorResponse.HttpErrorResponder(w)
			}
			if isValidInput(&sentRestaurant, &myErrorResponse) == false {
				fmt.Println("Something not valid")
				myErrorResponse.Code = http.StatusBadRequest
				myErrorResponse.HttpErrorResponder(w)
	
				return
			}
			
			dbRestaurant.Name = sentRestaurant.Name
			row := dbRestaurant.GetRestaurantInfoByName()
			if row == 0 {
				// no entry found in db
				//add review
			} else if row == -1 {
				//something bad happened
				fmt.Printf("something went while retrieving data %v", err)
				myErrorResponse.Code = http.StatusInternalServerError
				myErrorResponse.CustomMessage = "something went while retrieving data"
				myErrorResponse.HttpErrorResponder(w)
			} else if sentRestaurant.Source == dbRestaurant.Source {
			//same source, check location ID for same restaurnt
			if sentRestaurant.SourceLocID != dbRestaurant.SourceLocID {
				//creaet new restaurant with +1 to location_num
			} else {
				//use existing DB values
				//create review
			}
			} else if dbRestaurant.Source == "instagram"  {
					//trust DB over New
			} else if sentRestaurant.Source == "instagram" {
					//trust this guy
			}
			

			w.WriteHeader(http.StatusCreated)
			return

		case "GET":

			fmt.Println("Working Skel ")
			// w.Header().Set("Content-Type", "application/json")
			// json.NewEncoder(w).Encode(photoInfo)
			// if err != nil {
			//     http.Error(w, err.Error(), http.StatusInternalServerError)
			//     return
			// }
			return

		case "PUT":

			w.WriteHeader(http.StatusNoContent)
            return

		case "DELETE":

            w.WriteHeader(http.StatusNoContent)
            return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

func isValidInput(userInfo *db.Restaurants, errorResponse *globalsessionkeeper.ErrorResponse) bool {
	// if isValidString(userInfo.Email) == false {
	// 	fmt.Println("not valid email = ", userInfo.Email)
	// 	errorResponse.CustomMessage = "Invalid Email " + userInfo.Email
	// 	return false
	// }
	// if isValidString(userInfo.Username) == false {
	// 	fmt.Println("not valid username", userInfo.Username)
	// 	errorResponse.CustomMessage = "Invalid Username " + userInfo.Username
	// 	return false
	// }
	// if isValidString(userInfo.Password) == false {
	// 	fmt.Println("not valid password", userInfo.Password)
	// 	errorResponse.CustomMessage = "Invalid Password " + userInfo.Password
	// 	return false
	// }
	// if userInfo.Dob == 0 || age(time.Unix(int64(userInfo.Dob), 0)) < 18 {
	// 	errorResponse.CustomMessage = "Invalid Age " + string(age(time.Unix(int64(userInfo.Dob), 0)))
	// 	return false
	// }
	
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
