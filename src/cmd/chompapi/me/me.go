package me

import (
	"encoding/json"
	"fmt"
	"net/http"
	"cmd/chompapi/db"
	"cmd/chompapi/globalsessionkeeper"
	"reflect"
	"github.com/pborman/uuid"
	"github.com/gorilla/mux"
	"strconv"
	"strings"
	"github.com/parnurzeal/gorequest"
	"io/ioutil"
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

type InstagramTokenRequest struct {
	ClientId 				string 	`json:"client_id"`
	ClientSecret 			string 	`json:"client_secret"`
	RedirectUri 			string 	`json:"redirect_uri"`
	Code 					string 	`json:"code"`
}

type InstagramTokenReturn struct {
	AccessToken 			string `json:"access_token"`
	User 					InstaUser 	`json:"user"`

}

type InstaUser struct {
	Id						int 	`json:"id"`
	Username				string 	`json:"string"`
	FullName 				string 	`json:"full_name"`
	ProfilePicture 			string  `json:"profile_picture"`

}


func GetMe(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {

	userInfo := new(db.UserInfo)
	sessionUser := a.SessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))
	userInfo.Username = username

	switch r.Method {

	case "GET":

		err := userInfo.GetUserInfo(a.DB)
		if err != nil {
			//need logging here instead of print
			fmt.Println("Username not found..", userInfo.Username)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Username Not Found" + err.Error()}
		} else {
			fmt.Println("type for userInfo = ", userInfo)
			w.Header().Set("Content-Type", "application/json")
			userInfo.PasswordHash = ""
			json.NewEncoder(w).Encode(userInfo)
	
			if err != nil {
				return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Malformed JSON " + err.Error()}
			}
			return nil
		}
	default:

		fmt.Printf("Made it here.. method = %v\n", r.Method)
		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Method Not Allowed"}
	}
}

func PostPhotoId(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {

	userInfo := new(db.UserInfo)
	sessionUser := a.SessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))	
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))
	userInfo.Username = username
	fmt.Printf("Method = %v\n", r.Method)

	switch r.Method {

	case "POST":
		var photoInfo db.Photos
		w.Header().Set("Content-Type", "application/json")

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&photoInfo); err != nil {
			fmt.Printf("something %v\n", err)
			fmt.Printf("Photos = %v\n", photoInfo)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Malformed JSON:-:" +  err.Error()}
		}

		photoInfo.Uuid = GenerateUuid()
		photoInfo.Username = username
		fmt.Printf("photoInfo = %v\n", photoInfo)
	
		err := photoInfo.SetMePhoto(a.DB)
		if err != nil {
			//need logging here instead of print
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
		} 
		err2 := photoInfo.GetPhotoInfoByUuid(a.DB)
		if err2 != nil {
			//need logging here instead of print
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
		}

		w.Header().Set("Location", fmt.Sprintf("https://chompapi.com/me/photos/%v",  photoInfo.ID))
		w.Header().Set("UUID", photoInfo.Uuid)
		w.WriteHeader(http.StatusCreated)
		return nil

	case "GET":
		//variable definition
		var photoInfo db.Photos
		photoInfo.Username = username
		vars := mux.Vars(r)

    	photo_id, err := strconv.Atoi(vars["photoID"])
    	if err != nil {
    		fmt.Println("Not An Integer")
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Bad Photo ID " + err.Error()}
    	}
    	//collect photo ID
    	photoInfo.ID =  photo_id

        err = photoInfo.GetMePhotoByPhotoID(a.DB)
        if err != nil {
         	return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
        } else {
            fmt.Println("type for userInfo = ", photoInfo)
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(photoInfo)
            if err != nil {
            	return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
            }
            return nil
         }
         return nil

	case "PUT":
		//variable definition
		var photoInfo db.Photos

		vars := mux.Vars(r)
    	photo_id, err := strconv.Atoi(vars["photoID"])
    	if err != nil {
    		fmt.Println("Not An Integer")
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Bad Photo ID " + err.Error()}
    	}
    	//collect photo info and gen uuid
    	decoder := json.NewDecoder(r.Body)
		if err = decoder.Decode(&photoInfo); err != nil {
			fmt.Printf("something %v", err)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Malformed JSON:-:" +  err.Error()}
		}

		photoInfo.Username = username
    	photoInfo.ID =  photo_id

    	photoInfo.Uuid = GenerateUuid()
    	fmt.Println("uuid = ", photoInfo.Uuid)
    	if photoInfo.Uuid == "" {
    		return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
    	}
    	//add username to struct
		photoInfo.Username = username
	
         err = photoInfo.UpdateMePhoto(a.DB)
         if err != nil {
             //need logging here instead of print
         	return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
         } 
         err = photoInfo.UpdatePhotoIDUserTable(a.DB)
		if err != nil {
			//need logging here instead of print
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
		}

		w.Header().Set("Location", fmt.Sprintf("https://chompapi.com/me/photos/%v",  photoInfo.ID))
		w.Header().Set("UUID", photoInfo.Uuid)
		w.WriteHeader(http.StatusNoContent)
        return nil

	case "DELETE":
		//variable definition
		var photoInfo db.Photos
		photoInfo.Username = username
		vars := mux.Vars(r)

    	photo_id, err := strconv.Atoi(vars["photoID"])
    	if err != nil {
    		fmt.Println("Not An Integer")
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Bad Photo ID " + err.Error()}
    	}
    	//collect photo info
    	photoInfo.ID =  photo_id

        err = photoInfo.DeleteMePhoto(a.DB)
        if err != nil {
            //need logging here instead of print
        	return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
        }
        //change userid and update table
        photoInfo.ID = 0
        err = photoInfo.UpdatePhotoIDUserTable(a.DB)
		if err != nil {
			//need logging here instead of print
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
		}
		w.WriteHeader(http.StatusNoContent)
		return nil

	default:
		fmt.Printf("Made it here.. method = %v\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func DeleteMe(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {

	sessionUser := a.SessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))
	switch r.Method {

	case "DELETE":
		//variable definition
		var userInfo db.UserInfo
		userInfo.Username = username
		vars := mux.Vars(r)

		userId, err := strconv.Atoi(vars["userID"])
		if err != nil {
			fmt.Println("Not An Integer")
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Bad User ID " + err.Error()}
		}

		fmt.Printf("Getting user info for userid %v\n", userId)
		userInfo.UserID =  userId
		err = userInfo.GetUserInfo(a.DB)
		if err != nil {
	        //need logging here instead of print
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
	    }

	    fmt.Printf("Abandinging all photos for user %v\n", userInfo.Username)
	    err = userInfo.AbandonAllPhotos(a.DB)
		if err != nil {
			//need logging here instead of print
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
		}
		
	    //change userid and update table
		fmt.Printf("Deleting me for user %v, photo ID = %v\n", userInfo.Username, userInfo.Photo.ID)
		fmt.Printf("Deleting me photo %v\n", userInfo.Photo.ID)
	    photoInfo := new(db.Photos)
	    photoInfo.ID = userInfo.Photo.ID
	    if userInfo.Photo.ID != 0 {
	    	err = photoInfo.DeleteMePhoto(a.DB)
			if err != nil {
				//need logging here instead of print
				return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
			}
	    }

	    fmt.Printf("Deleting reviews for user %v\n", userInfo.Username)
		err = userInfo.DeleteAllReviewsByUser(a.DB)
	    if err != nil {
	        //need logging here instead of print
	        if strings.Contains("0 rows deleted", err.Error()) == false  {
	        	return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
	        }
	    }

	    fmt.Printf("Deleting crawl info for user %v\n", userInfo.Username)
		err = userInfo.DeleteCrawlInfo(a.DB)
	    if err != nil {
	        //need logging here instead of print
	        if strings.Contains("0 rows deleted", err.Error()) == false  {
	        	return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
	        }
	    }

	    fmt.Printf("Deleting user %v\n", userInfo.Username)
	    err = userInfo.DeleteUser(a.DB)
	    if err != nil {
	        //need logging here instead of print
	    	return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
	    }

		// err = userInfo.DeleteAllPhotos()
		// if err != nil {
		// 	//need logging here instead of print
		// 	myErrorResponse.Code = http.StatusInternalServerError
		// 	myErrorResponse.Desc= err.Error()
		// 	myErrorResponse.HttpErrorResponder(w)
		// 	return
		// }
		fmt.Printf("Logging all sessions out for user %v\n", userInfo.Username)
		err = db.LogoutAllSessions(userInfo.Username, a.DB)

		if err != nil {
			//need logging here instead of print
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
		}
	    w.WriteHeader(http.StatusNoContent)
	    return nil

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func Logout(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {

	userInfo := new(db.UserInfo)
	cookie := globalsessionkeeper.GetCookie(r)
	sessionUser := a.SessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))
	userInfo.Username = username

	switch r.Method {

	case "POST":

		err := db.Logout(cookie, a.DB)
		if err != nil {
			//need logging here instead of print
			fmt.Println("Username not found..", userInfo.Username)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Username Not Found" + err.Error()}
		} else {
			fmt.Printf("Logged out user %v, sessionId = %v\n", userInfo.Username, cookie)
			w.Header().Set("Content-Type", "application/json")
			err = globalsessionkeeper.ExpireCookie(r, w)
			if err != nil {
				fmt.Printf("Error = %v\n", err)
			}
			w.WriteHeader(http.StatusNoContent)
			return nil
		}
	default:

		fmt.Printf("Made it here.. method = %v\n", r.Method)
		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Method Not Allowed"}
	}
}

func LogoutAll(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {

	userInfo := new(db.UserInfo)
	cookie := globalsessionkeeper.GetCookie(r)
	sessionUser := a.SessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))
	userInfo.Username = username

	switch r.Method {

	case "POST":

		err := db.LogoutAllSessions(username, a.DB)
		if err != nil {
			//need logging here instead of print
			fmt.Println("Username not found..", userInfo.Username)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Username Not Found" + err.Error()}
		} else {
			fmt.Printf("Logged out user %v, sessionId = %v\n", userInfo.Username, cookie)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNoContent)
			return nil
		}
	default:

		fmt.Printf("Made it here.. method = %v\n", r.Method)
		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Method Not Allowed"}
	}
}

func Instagram(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {

	userInfo := new(db.UserInfo)
	sessionUser := a.SessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))
	userInfo.Username = username

	switch r.Method {

	case "GET":
		w.Header().Set("Content-Type", "application/json")
		//instagram config 
		instaConfig := new(InstagramTokenRequest)
		instaConfigFile, err := ioutil.ReadFile("./chomp_private/instagram_auth.conf")
		if err != nil {
		    return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, err.Error()}
		}
		err = json.Unmarshal(instaConfigFile, &instaConfig)
		if err != nil {
		    fmt.Printf("Err = %v", err)
		    return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, err.Error()}
		}
		query := mux.Vars(r)
		fmt.Printf("Query %v\n", query)
		if query["error"] != "" {
			fmt.Printf("Error not nil, updating error instacode %v\n", query["error"])
			userInfo.InstaCode = ""
			err := userInfo.UpdateInstaCode(a.DB)
			if err != nil {
				fmt.Printf("Err updating 1 instacode %v\n", err)
				return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, err.Error()}
			}
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, query["error"]}
		}
		instaConfig.Code = query["code"]
		userInfo.InstaToken, err = getInstagramToken(instaConfig)
		if err != nil {
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, err.Error()}
		}
		userInfo.InstaCode = query["code"]
		err = userInfo.UpdateInstaCode(a.DB)
		if err != nil {
			fmt.Printf("Err updating 2 instacode %v\n", err)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, err.Error()}
		}
		return nil
		
	default:

		fmt.Printf("Made it here.. method = %v\n", r.Method)
		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Method Not Allowed"}
	}
}

func getInstagramToken(instagramTokenReq *InstagramTokenRequest) (string, error) {
	iurl :=  "https://api.instagram.com/oauth/access_token"
	// instagramTokenReq := new(InstagramTokenRequest)

	request := gorequest.New()
	resp, body, errs := request.Post(iurl).Send(instagramTokenReq).End()
	if errs != nil {
		fmt.Printf("something went wrong in get %v", errs)
	}
	fmt.Printf("Body = %v\n", body)
	fmt.Printf("resp = %v\n", resp)
	return "access_token", errs[0]
}

func InstagramGetAccessToken(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {

	userInfo := new(db.UserInfo)
	sessionUser := a.SessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))
	userInfo.Username = username

	switch r.Method {

	case "GET":
		w.Header().Set("Content-Type", "application/json")
		// query := mux.Vars(r)
		// fmt.Printf("Query %v\n", query)
		// if query["error"] != "" {
		// 	fmt.Printf("Error not nil, updating error instacode %v\n", query["error"])
		// 	userInfo.InstaCode = ""
		// 	err := userInfo.UpdateInstaCode(a.DB)
		// 	if err != nil {
		// 		fmt.Printf("Err updating 1 instacode %v\n", err)
		// 		return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, err.Error()}
		// 	}
		// 	return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, query["error"]}
		// }
		// userInfo.InstaCode = query["code"]
		err := userInfo.GetInstagramAccessToken(a.DB)
		if err != nil {
			fmt.Printf("Err getting instagram token %v\n", err)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, err.Error()}
		}
		return nil
		
	default:

		fmt.Printf("Non supported method called in InstagramGetAccessToken = %v\n", r.Method)
		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Method Not Allowed"}
	}
}


func InstagramLinkClick(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {
	sessionUser := a.SessionStore.Get("username")
	username := reflect.ValueOf(sessionUser).String()
	fmt.Printf("Found Session! Session username = %v\n", sessionUser)
	fmt.Printf("values = %v\n", reflect.TypeOf(sessionUser))
	switch r.Method {
	case "PUT":

		dbUserInfo := new(db.UserInfo)
		dbUserInfo.Username = username
		err := dbUserInfo.GetUserInfo(a.DB)
		if err != nil {
			fmt.Printf("Failed to get userinfo, err = %v\n", err)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, err.Error()}
		}
		dbUserInfo.Username = username

		fmt.Printf("Json Input = %+v\n", dbUserInfo)
		fmt.Printf("pass = %v\n", dbUserInfo.Password)

		err = dbUserInfo.InstagramLinkClick(a.DB)
		if err != nil {
			fmt.Println("Something not valid")
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, err.Error()}
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
		
	default:

		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Method Not Allowed"}
	}
}

func DeactivateMe(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {

	switch r.Method {

	default:
		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Method Not Allowed"}
	}
}

func GenerateUuid() string {
	myUuid := uuid.NewRandom()
	return myUuid.String()
}
