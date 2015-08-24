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


func Reviews(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {

	sessionUser := a.SessionStore.Get("username")
	sessionUserID := a.SessionStore.Get("userId")
	fmt.Println("SessionUser = %v", sessionUser)
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
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Malformed JSON: " + err.Error()}
		}

		fmt.Printf("Dishtags = %v\n", review.DishTags)
		fmt.Printf("Review = %v\n", review)
		dbRestaurant.Name = review.Restaurant.Name
		err := dbRestaurant.GetRestaurantInfoByName(a.DB)
		if err != nil && err != sql.ErrNoRows{
			//something bad happened
			fmt.Printf("something went while retrieving data %v", err)
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "something went while retrieving data:-:" + err.Error()}
		} else if err == sql.ErrNoRows || dbRestaurant.ID == 0 {
			// not found in DB
			if review.Restaurant.Name != "" {
				fmt.Println("Restaurant Not found in DB, creating new entry")
				err = review.Restaurant.CreateRestaurant(a.DB)
				if err != nil {
					//something bad happened
					fmt.Printf("something went while retrieving data %v", err)
					return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "something went while retrieving data:-:" + err.Error()}
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
					err = review.Restaurant.CreateRestaurant(a.DB)
					if err != nil {
						//something bad happened
						fmt.Printf("something went while retrieving data %v", err)
						return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "something went while retrieving data:-:" + err.Error()}
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
					review.Restaurant.UpdateRestaurant(a.DB)
					if err != nil {
						//something bad happened
						fmt.Printf("something went while retrieving data %v", err)
						return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "something went while retrieving data:-:" + err.Error()}
					}
				} else {
					fmt.Println("location id !=")
					review.Restaurant.LocationNum = dbRestaurant.LocationNum + 1
					err = review.Restaurant.CreateRestaurant(a.DB)
					if err != nil {
						//something bad happened
						fmt.Printf("something went while retrieving data %v", err)
						return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "something went while retrieving data:-:" + err.Error()}
					}
				}
			}  
		}  
		// all other cases
		//Validate dish
		dbDish.Name = review.Dish.Name
		err = dbDish.GetDishInfoByName(a.DB)
		if err != nil && err != sql.ErrNoRows{
			//something bad happened
			fmt.Printf("something went while retrieving data %v", err)
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "something went while retrieving data:-:" + err.Error()}
		} else if err == sql.ErrNoRows {
			// not found in DB
			fmt.Println("Not found in DB, creating new entry")
			err = review.Dish.CreateDish(a.DB)
			if err != nil {
				//something bad happened
				fmt.Printf("something went while retrieving data %v", err)
				return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "something went while retrieving data:-:" + err.Error()}
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
				return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Invalid Review ID"}
   			}
   			review.ID = review_id
			err = review.UpdateReview(a.DB)
			if err != nil {
				//something bad happened
				if err.Error() == "0 rows updated" {
					fmt.Printf("something went while retrieving data %v", err)
					return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, err.Error()}
				} else {
					fmt.Printf("something went while retrieving data %v", err)
					return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "could not update review:-:" + err.Error()}
				}
			} else {
				review.GetReviewLastTimeStamp(review.ID, a.DB)
				lastUpdate := strconv.Itoa(review.LastUpdated)
				w.Header().Set("LastUpdated", lastUpdate)
				w.WriteHeader(http.StatusNoContent)
			}
		} else {
			err = review.CreateReview(a.DB)
			if err != nil {
				//something bad happened
				fmt.Printf("something went while retrieving data %v", err)
				return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "could not create review:-:" + err.Error()}
			} else {
				review.GetReviewLastTimeStamp(review.ID, a.DB)
				lastUpdate := strconv.Itoa(review.LastUpdated)
				w.Header().Set("LastUpdated", lastUpdate)
				w.Header().Set("Location", fmt.Sprintf("https://chompapi.com/reviews/%d", review.ID))
				w.WriteHeader(http.StatusCreated)
			}
		}
		// return
		return nil

	case "GET":

		fmt.Println("Working Skel ")
		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Method Not Allowed"}

	case "DELETE":

	vars := mux.Vars(r)
   	review_id, thisErr := strconv.Atoi(vars["reviewID"])
   	if thisErr != nil {
   		fmt.Println("Not An Integer")
		return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Invalid Review ID"}
   	}
   	review.ID = review_id
		err := review.DeleteReview(a.DB)
		if err != nil {
			//something bad happened
			if err.Error() == "0 rows deleted" {
				fmt.Printf("something went while retrieving data %v", err)
				return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Error: " + err.Error()}
			} else {
				fmt.Printf("something went while retrieving data %v", err)
				return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "could not create review"}
			}
		}
       w.WriteHeader(http.StatusNoContent)
       return nil

	default:
		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Method Not Allowed"}
	}
}
