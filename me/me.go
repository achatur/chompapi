package me

import (
	"encoding/json"
	"fmt"
	"net/http"
	"chompapi/db"
	"chompapi/globalsessionkeeper"
	"reflect"
	"github.com/pborman/uuid"
	"github.com/gorilla/mux"
	"strconv"
	"strings"
)

type UserInfo struct {
	ID   			string
	Username 		string
	Email         	string
	DOB           	string
	Gender        	string
	Photo		  	Photo
}

type Photo struct {
	ID			string 		`json:"id"`
}


func GetMe(w http.ResponseWriter, r *http.Request) {

	userInfo := new(db.UserInfo)
	var myErrorResponse globalsessionkeeper.ErrorResponse
	cookie := globalsessionkeeper.GetCookie(r)
	sessionStore, err := globalsessionkeeper.GlobalSessions.GetSessionStore(cookie)
	if err != nil {
		//need logging here instead of print
		myErrorResponse.Code = http.StatusUnauthorized
		myErrorResponse.Error = err.Error()
		myErrorResponse.HttpErrorResponder(w)
		return
	}
	sessionUser := sessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	//need logging here instead of print
	//extend session time by GC time
	defer sessionStore.SessionRelease(w)
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))
	userInfo.Username = username

	switch r.Method {

	case "GET":

		err = userInfo.GetUserInfo()
		if err != nil {
			//need logging here instead of print
			fmt.Println("Username not found..", userInfo.Username)
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "Username Not Found" + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		} else {
			fmt.Println("type for userInfo = ", userInfo)
			w.Header().Set("Content-Type", "application/json")
			userInfo.PasswordHash = ""
			json.NewEncoder(w).Encode(userInfo)
	
			if err != nil {
				myErrorResponse.Code = http.StatusBadRequest
				myErrorResponse.Error = "Malformed JSON " + err.Error()
				myErrorResponse.HttpErrorResponder(w)
				return
			}
			return
		}
	default:

		fmt.Printf("Made it here.. method = %v\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func PostPhotoId(w http.ResponseWriter, r *http.Request) {

	userInfo := new(db.UserInfo)
	var myErrorResponse globalsessionkeeper.ErrorResponse
	cookie := globalsessionkeeper.GetCookie(r)
	sessionStore, err := globalsessionkeeper.GlobalSessions.GetSessionStore(cookie)
	if err != nil {
		//need logging here instead of print
		myErrorResponse.Code = http.StatusUnauthorized
		myErrorResponse.Error = err.Error()
		myErrorResponse.HttpErrorResponder(w)
		return
	}
	sessionUser := sessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	//need logging here instead of print
	//extend session time by GC time
	defer sessionStore.SessionRelease(w)
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))
	userInfo.Username = username
	// defer sessionStore.SessionRelease(w)
	// username := reflect.ValueOf(sessionUser).String()
	fmt.Printf("Method = %v\n", r.Method)
	switch r.Method {

	case "POST":
		var photoInfo db.Photos
		w.Header().Set("Content-Type", "application/json")

		photoInfo.Uuid = GenerateUuid()
		photoInfo.Username = username
	
		err := photoInfo.SetMePhoto()
		if err != nil {
			//need logging here instead of print
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		} 
		err2 := photoInfo.GetPhotoInfoByUuid()
		if err2 != nil {
			//need logging here instead of print
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		w.Header().Set("Location", fmt.Sprintf("https://chompapi.com/me/photos/%v",  photoInfo.ID))
		w.Header().Set("UUID", photoInfo.Uuid)
		w.WriteHeader(http.StatusCreated)
		return

	case "GET":
		//variable definition
		var photoInfo db.Photos
		photoInfo.Username = username
		vars := mux.Vars(r)

    	photo_id, err := strconv.Atoi(vars["photoID"])
    	if err != nil {
    		fmt.Println("Not An Integer")
    		myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "Bad Photo ID " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
    	}
    	//collect photo ID
    	photoInfo.ID =  photo_id

         err = photoInfo.GetMePhotoByPhotoID()
         if err != nil {
             //need logging here instead of print
             myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
             return
         } else {
             fmt.Println("type for userInfo = ", photoInfo)
             w.Header().Set("Content-Type", "application/json")
             json.NewEncoder(w).Encode(photoInfo)
             if err != nil {
                 myErrorResponse.Code = http.StatusInternalServerError
				myErrorResponse.Error = err.Error()
				myErrorResponse.HttpErrorResponder(w)
                 return
             }
             return
         }
         return

	case "PUT":
		//variable definition
		var photoInfo db.Photos
		photoInfo.Username = username

		vars := mux.Vars(r)
    	photo_id, thisErr := strconv.Atoi(vars["photoID"])
    	if thisErr != nil {
    		fmt.Println("Not An Integer")
    		myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "Bad Photo ID " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
    	}
    	//collect photo info and gen uuid
    	photoInfo.ID =  photo_id

    	photoInfo.Uuid = GenerateUuid()
    	fmt.Println("uuid = ", photoInfo.Uuid)
    	if photoInfo.Uuid == "" {
    		myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
    		return
    	}
    	//add username to struct
		photoInfo.Username = username
	
         err := photoInfo.UpdateMePhoto()
         if err != nil {
             //need logging here instead of print
             myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
             return
         } 
         err = photoInfo.UpdatePhotoIDUserTable()
		if err != nil {
			//need logging here instead of print
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		w.Header().Set("Location", fmt.Sprintf("https://chompapi.com/me/photos/%v",  photoInfo.ID))
		w.Header().Set("UUID", photoInfo.Uuid)
		w.WriteHeader(http.StatusNoContent)
         return

	case "DELETE":
		//variable definition
		var photoInfo db.Photos
		photoInfo.Username = username
		vars := mux.Vars(r)

    	photo_id, thisErr := strconv.Atoi(vars["photoID"])
    	if thisErr != nil {
    		fmt.Println("Not An Integer")
    		myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "Bad Photo ID " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
    	}
    	//collect photo info
    	photoInfo.ID =  photo_id

         err := photoInfo.DeleteMePhoto()
         if err != nil {
             //need logging here instead of print
             myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
             return
         }
         //change userid and update table
         photoInfo.ID = 0
         err = photoInfo.UpdatePhotoIDUserTable()
		if err != nil {
			//need logging here instead of print
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		fmt.Printf("Made it here.. method = %v\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func DeleteMe(w http.ResponseWriter, r *http.Request) {

	var myErrorResponse globalsessionkeeper.ErrorResponse
	cookie := globalsessionkeeper.GetCookie(r)
	sessionStore, err := globalsessionkeeper.GlobalSessions.GetSessionStore(cookie)
	if err != nil {
		//need logging here instead of print
		myErrorResponse.Code = http.StatusUnauthorized
		myErrorResponse.Error = err.Error()
		myErrorResponse.HttpErrorResponder(w)
		return
	}
	sessionUser := sessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	switch r.Method {

	case "DELETE":
		//variable definition
		var userInfo db.UserInfo
		userInfo.Username = username
		vars := mux.Vars(r)

		userId, err := strconv.Atoi(vars["userID"])
		if err != nil {
			fmt.Println("Not An Integer")
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "Bad User ID " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		fmt.Printf("Getting user info for userid %v\n", userId)
		userInfo.UserID =  userId
		err = userInfo.GetUserInfo()
		if err != nil {
	        //need logging here instead of print
	        myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
	        return
	    }

	    fmt.Printf("Deleting reviews for user %v\n", userInfo.Username)
		err = userInfo.DeleteAllReviewsByUser()
	    if err != nil {
	        //need logging here instead of print
	        if strings.Contains("0 rows deleted", err.Error()) == false  {

	        	myErrorResponse.Code = http.StatusInternalServerError
				myErrorResponse.Error = err.Error()
				myErrorResponse.HttpErrorResponder(w)
	        	return
	        }
	    }

	    fmt.Printf("Abandinging all photos for user %v\n", userInfo.Username)
	    err = userInfo.AbandonAllPhotos()
		if err != nil {
			//need logging here instead of print
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		//change userid and update table
		fmt.Printf("Deleting me for user %v, photo ID = %v\n", userInfo.Username, userInfo.Photo.ID)
		fmt.Printf("Deleting me photo %v\n", userInfo.Photo.ID)
	    photoInfo := new(db.Photos)
	    photoInfo.ID = userInfo.Photo.ID
	    err = photoInfo.DeleteMePhoto()
		if err != nil {
			//need logging here instead of print
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}

	    fmt.Printf("Deleting user %v\n", userInfo.Username)
	    err = userInfo.DeleteUser()
	    if err != nil {
	        //need logging here instead of print
	        myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
	        return
	    }

		// err = userInfo.DeleteAllPhotos()
		// if err != nil {
		// 	//need logging here instead of print
		// 	myErrorResponse.Code = http.StatusInternalServerError
		// 	myErrorResponse.Error = err.Error()
		// 	myErrorResponse.HttpErrorResponder(w)
		// 	return
		// }
		fmt.Printf("Logging all sessions out for user %v\n", userInfo.Username)
		err = db.LogoutAllSessions(userInfo.Username)
		if err != nil {
			//need logging here instead of print
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}
	    w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {

	userInfo := new(db.UserInfo)
	var myErrorResponse globalsessionkeeper.ErrorResponse
	cookie := globalsessionkeeper.GetCookie(r)
	sessionStore, err := globalsessionkeeper.GlobalSessions.GetSessionStore(cookie)
	if err != nil {
		//need logging here instead of print
		myErrorResponse.Code = http.StatusUnauthorized
		myErrorResponse.Error = err.Error()
		myErrorResponse.HttpErrorResponder(w)
		return
	}
	sessionUser := sessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	//need logging here instead of print
	//extend session time by GC time
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))
	userInfo.Username = username

	switch r.Method {

	case "POST":

		err = db.Logout(cookie)
		if err != nil {
			//need logging here instead of print
			fmt.Println("Username not found..", userInfo.Username)
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "Username Not Found" + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		} else {
			fmt.Printf("Logged out user %v, sessionId = %v\n", userInfo.Username, cookie)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNoContent)
			return
		}
	default:

		fmt.Printf("Made it here.. method = %v\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func LogoutAll(w http.ResponseWriter, r *http.Request) {

	userInfo := new(db.UserInfo)
	var myErrorResponse globalsessionkeeper.ErrorResponse
	cookie := globalsessionkeeper.GetCookie(r)
	sessionStore, err := globalsessionkeeper.GlobalSessions.GetSessionStore(cookie)
	if err != nil {
		//need logging here instead of print
		myErrorResponse.Code = http.StatusUnauthorized
		myErrorResponse.Error = err.Error()
		myErrorResponse.HttpErrorResponder(w)
		return
	}
	sessionUser := sessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	//need logging here instead of print
	//extend session time by GC time
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))
	userInfo.Username = username

	switch r.Method {

	case "POST":

		err = db.LogoutAllSessions(username)
		if err != nil {
			//need logging here instead of print
			fmt.Println("Username not found..", userInfo.Username)
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "Username Not Found" + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		} else {
			fmt.Printf("Logged out user %v, sessionId = %v\n", userInfo.Username, cookie)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNoContent)
			return
		}
	default:

		fmt.Printf("Made it here.. method = %v\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func Instagram(w http.ResponseWriter, r *http.Request) {

	userInfo := new(db.UserInfo)
	var myErrorResponse globalsessionkeeper.ErrorResponse
	cookie := globalsessionkeeper.GetCookie(r)
	sessionStore, err := globalsessionkeeper.GlobalSessions.GetSessionStore(cookie)
	if err != nil {
		//need logging here instead of print
		myErrorResponse.Code = http.StatusUnauthorized
		myErrorResponse.Error = err.Error()
		myErrorResponse.HttpErrorResponder(w)
		return
	}
	sessionUser := sessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	//need logging here instead of print
	//extend session time by GC time
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))
	userInfo.Username = username

	switch r.Method {

	case "GET":
		w.Header().Set("Content-Type", "application/json")
		query := mux.Vars(r)
		fmt.Printf("Query %v\n", query)
		if query["error"] != "" {
			fmt.Printf("Error not nil, updating error instacode %v\n", query["error"])
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = query["error"]
			myErrorResponse.HttpErrorResponder(w)
			userInfo.InstaCode = ""
			err = userInfo.UpdateInstaCode()
			if err != nil {
				fmt.Printf("Err updating 1 instacode %v\n", err)
				myErrorResponse.Code = http.StatusBadRequest
				myErrorResponse.Error = err.Error()
				myErrorResponse.HttpErrorResponder(w)
			}
			return
		}
		userInfo.InstaCode = query["code"]
		err = userInfo.UpdateInstaCode()
		if err != nil {
			fmt.Printf("Err updating 2 instacode %v\n", err)
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		return
		
	default:

		fmt.Printf("Made it here.. method = %v\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func InstagramLinkClick(w http.ResponseWriter, r *http.Request) {
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
	fmt.Printf("This SessionId = %v\n", sessionUserID)


	defer sessionStore.SessionRelease(w)
	//create variables
	username := reflect.ValueOf(sessionUser).String()
	switch r.Method {
	case "PUT":

		// input := new(db.UserInfo)
		dbUserInfo := new(db.UserInfo)
		dbUserInfo.Username = username
		err = dbUserInfo.GetUserInfo()
		if err != nil {
			fmt.Printf("Failed to get userinfo, err = %v\n", err)
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		dbUserInfo.Username = username

		fmt.Printf("Json Input = %+v\n", dbUserInfo)
		fmt.Printf("pass = %v\n", dbUserInfo.Password)

		err = dbUserInfo.InstagramLinkClick()

		if err != nil {
			fmt.Println("Something not valid")
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
		
	default:

		myErrorResponse.Code = http.StatusMethodNotAllowed
		myErrorResponse.Error = "Invalid Method"
		myErrorResponse.HttpErrorResponder(w)
		return
	}
}

func DeactivateMe(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func GenerateUuid() string {
	myUuid := uuid.NewRandom()
	return myUuid.String()
}
