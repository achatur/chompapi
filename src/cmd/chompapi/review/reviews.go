package review

import (
	"encoding/json"
	"fmt"
	"net/http"
	"cmd/chompapi/db"
	"cmd/chompapi/globalsessionkeeper"
	"reflect"
	"database/sql"
	"github.com/gorilla/mux"
	"strconv"
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

	sessionUser := sessionStore.Get("username")
	sessionUserID := sessionStore.Get("userId")
	fmt.Println("SessionUser = %v", sessionUser)

	//reset time to time.now() + maxlifetime
	defer sessionStore.SessionRelease(w)

	//create variables
	username 	 := reflect.ValueOf(sessionUser).String()
	userId 	 	 := reflect.ValueOf(sessionUserID).Int()
	review 	 	 := new(db.Review)
	dbRestaurant := new(db.Restaurants)
	dbDish 		 := new(db.Dish)

	review.Username = username
	review.UserID = int(userId)

	switch r.Method {

	case "PUT", "POST":
		
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&review); err != nil {
			//need logging here instead of print
			fmt.Printf("something went wrong in reviews %v\n", err.Error())
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "Malformed JSON: " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		fmt.Printf("Dishtags = %v\n", review.DishTags)
		fmt.Printf("Review = %v\n", review)
		dbRestaurant.Name = review.Restaurant.Name
		err2 := dbRestaurant.GetRestaurantInfoByName()
		if err2 != nil && err2 != sql.ErrNoRows{
			//something bad happened
			fmt.Printf("something went while retrieving data %v", err)
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = "something went while retrieving data:-:" + err2.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		} else if err2 == sql.ErrNoRows || dbRestaurant.ID == 0 {
			// not found in DB
			if review.Restaurant.Name != "" {
				fmt.Println("Restaurant Not found in DB, creating new entry")
				err = review.Restaurant.CreateRestaurant()
				if err != nil {
					//something bad happened
					fmt.Printf("something went while retrieving data %v", err)
					myErrorResponse.Code = http.StatusInternalServerError
					myErrorResponse.Error = "something went while retrieving data:-:" + err.Error()
					myErrorResponse.HttpErrorResponder(w)	
					return
				}
			} else {
				// Restaurant Value Blank
				fmt.Println("Blank Restaurant found in db")
				fmt.Println("Blank Restaurant In DB", dbRestaurant)
				review.Restaurant = *dbRestaurant
			
			}
		} else {
			// entry found in db
			fmt.Println("Restaurant found in db")
			fmt.Println("Restaurant In DB", dbRestaurant)
			if review.Restaurant.Source == dbRestaurant.Source {
				//same source, check location ID for same restaurnt
				fmt.Println("same source")
				if review.Restaurant.SourceLocID != dbRestaurant.SourceLocID {
					//creaet new restaurant with +1 to location_num
					fmt.Println("location id !=")
					review.Restaurant.LocationNum = dbRestaurant.LocationNum + 1
					err = review.Restaurant.CreateRestaurant()
					if err != nil {
						//something bad happened
						fmt.Printf("something went while retrieving data %v", err)
						myErrorResponse.Code = http.StatusInternalServerError
						myErrorResponse.Error = "something went while retrieving data:-:" + err.Error()
						myErrorResponse.HttpErrorResponder(w)
						return	
					}
				} else {
					//use existing DB values
					fmt.Println("Location ID Equal, using db values")
					review.Restaurant = *dbRestaurant
				}
			} else if dbRestaurant.Source == "factual"  {
				//trust DB over New
				fmt.Println("Source not same, DB == factual")
				review.Restaurant = *dbRestaurant
				//review.CreateReview()
			} else if dbRestaurant.Source == "instagram" && review.Restaurant.Source != "factual"  {
				//trust DB over New
				fmt.Println("Source not same, DB == insta")
				review.Restaurant = *dbRestaurant
				//review.CreateReview()
			} else if review.Restaurant.Source == "instagram" ||
					   review.Restaurant.Source == "factual" {
				fmt.Printf("New restaurant %v, updating db\n", review.Restaurant.Source)
				if dbRestaurant.LocationNum == 0 {
					review.Restaurant.ID = dbRestaurant.ID
					review.Restaurant.UpdateRestaurant()
					if err != nil {
						//something bad happened
						fmt.Printf("something went while retrieving data %v", err)
						myErrorResponse.Code = http.StatusInternalServerError
						myErrorResponse.Error = "something went while retrieving data:-:" + err.Error()
						myErrorResponse.HttpErrorResponder(w)
						return	
					}
				} else {
					fmt.Println("location id !=")
					review.Restaurant.LocationNum = dbRestaurant.LocationNum + 1
					err = review.Restaurant.CreateRestaurant()
					if err != nil {
						//something bad happened
						fmt.Printf("something went while retrieving data %v", err)
						myErrorResponse.Code = http.StatusInternalServerError
						myErrorResponse.Error = "something went while retrieving data:-:" + err.Error()
						myErrorResponse.HttpErrorResponder(w)
						return	
					}
				}
			}  
		}  
		// all other cases
		//Validate dish
		dbDish.Name = review.Dish.Name
		err3 := dbDish.GetDishInfoByName()
		if err3 != nil && err3 != sql.ErrNoRows{
			//something bad happened
			fmt.Printf("something went while retrieving data %v", err)
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = "something went while retrieving data:-:" + err3.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		} else if err3 == sql.ErrNoRows {
			// not found in DB
			fmt.Println("Not found in DB, creating new entry")
			err = review.Dish.CreateDish()
			if err != nil {
				//something bad happened
				fmt.Printf("something went while retrieving data %v", err)
				myErrorResponse.Code = http.StatusInternalServerError
				myErrorResponse.Error = "something went while retrieving data:-:" + err.Error()
				myErrorResponse.HttpErrorResponder(w)
				return
			}
		} else {
			fmt.Println("Found Dish ", dbDish)
			review.Dish = *dbDish
		}
		fmt.Println("writing to db!")

		if r.Method == "PUT" {
			vars := mux.Vars(r)
   		review_id, thisErr := strconv.Atoi(vars["reviewID"])
   		if thisErr != nil {
   			fmt.Println("Not An Integer")
   			myErrorResponse.Code = http.StatusBadRequest
				myErrorResponse.Error = "Invalid Review ID"
				myErrorResponse.HttpErrorResponder(w)
			return
   		}
   		review.ID = review_id
			err = review.UpdateReview()
			if err != nil {
				//something bad happened
				if err.Error() == "0 rows updated" {
					fmt.Printf("something went while retrieving data %v", err)
					myErrorResponse.Code = http.StatusBadRequest
					myErrorResponse.Error = err.Error()
					myErrorResponse.HttpErrorResponder(w)
				} else {
					fmt.Printf("something went while retrieving data %v", err)
					myErrorResponse.Code = http.StatusInternalServerError
					myErrorResponse.Error = "could not update review:-:" + err.Error()
					myErrorResponse.HttpErrorResponder(w)
				}
			} else {
				review.GetReviewLastTimeStamp(review.ID)
				lastUpdate := strconv.Itoa(review.LastUpdated)
				w.Header().Set("LastUpdated", lastUpdate)
				w.WriteHeader(http.StatusNoContent)
			}
		} else {
			err = review.CreateReview()
			if err != nil {
				//something bad happened
				fmt.Printf("something went while retrieving data %v", err)
				myErrorResponse.Code = http.StatusBadRequest
				myErrorResponse.Error = "could not create review:-:" + err.Error()
				myErrorResponse.HttpErrorResponder(w)
				return
			} else {
				review.GetReviewLastTimeStamp(review.ID)
				review.GetReviewLastTimeStamp(review.ID)
				lastUpdate := strconv.Itoa(review.LastUpdated)
				w.Header().Set("LastUpdated", lastUpdate)
				w.Header().Set("Location", fmt.Sprintf("https://chompapi.com/reviews/%d", review.ID))
				w.WriteHeader(http.StatusCreated)
			}
		}
		return

	case "GET":

		fmt.Println("Working Skel ")
		return

	case "DELETE":

	vars := mux.Vars(r)
   	review_id, thisErr := strconv.Atoi(vars["reviewID"])
   	if thisErr != nil {
   		fmt.Println("Not An Integer")
   		myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "Invalid Review ID"
			myErrorResponse.HttpErrorResponder(w)
		return
   	}
   	review.ID = review_id
		err = review.DeleteReview()
		if err != nil {
			//something bad happened
			if err.Error() == "0 rows deleted" {
				fmt.Printf("something went while retrieving data %v", err)
				myErrorResponse.Code = http.StatusBadRequest
				myErrorResponse.Error = "Error: " + err.Error()
				myErrorResponse.HttpErrorResponder(w)
			} else {
				fmt.Printf("something went while retrieving data %v", err)
				myErrorResponse.Code = http.StatusInternalServerError
				myErrorResponse.Error = "could not create review"
				myErrorResponse.HttpErrorResponder(w)
			}
			
			return
		}
       w.WriteHeader(http.StatusNoContent)
       return

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
