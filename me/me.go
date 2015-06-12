package me

import (
	"encoding/json"
	"fmt"
	"net/http"
	"chompapi/db"
	"chompapi/globalsessionkeeper"
	"chompapi/login"
	"strings"
	"reflect"
	"github.com/pborman/uuid"
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
	ID			string
}


func GetMe(w http.ResponseWriter, r *http.Request) {

	var input login.LoginInput
	var returnUser UserInfo
	fmt.Printf("Number of active sessions: %v\n", globalsessionkeeper.GlobalSessions.GetActiveSession())
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
			//need logging here instead of print
		fmt.Printf("Found Session! Session username = %v\n", sessionUser)
		fmt.Printf("Found Session! Session username values = %v\n", reflect.TypeOf(sessionUser))
		input.Username = reflect.ValueOf(sessionUser).String()
		userInfo, err := db.GetMeInfo(input.Username)
		if err != nil {
			//need logging here instead of print
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		} else {
			fmt.Println("type for userInfo = ", userInfo)
			w.Header().Set("Content-Type", "application/json")
			returnUser.ID = userInfo["chomp_user_id"]
			returnUser.Username = userInfo["chomp_username"]
			returnUser.DOB = userInfo["dob"]
			returnUser.Email = userInfo["email"]
			returnUser.Photo.ID = userInfo["photo_id"]
			json.NewEncoder(w).Encode(returnUser)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
	}
}

func PostPhotoId(w http.ResponseWriter, r *http.Request) {

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
			var photoInfo db.Photos
			w.Header().Set("Content-Type", "application/json")
			
			photoInfo.Uuid = generateUuid()
			photoInfo.Username = username
		
			err := photoInfo.SetMePhoto()
			if err != nil {
				//need logging here instead of print
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			} 
			photoTable, err2 := db.GetPhotoInfoByUuid(photoInfo.Uuid)
			if err2 != nil {
				//need logging here instead of print
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			} 
		
			w.Header().Set("Location", fmt.Sprintf("https://chompapi.com/me/photos/%v",  photoTable["id"]))
			w.Header().Set("UUID", photoTable["uuid"])
			w.WriteHeader(http.StatusCreated)
			return
		case "GET":
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

func generateUuid() string {
	myUuid := uuid.NewRandom()
	return myUuid.String()
}

func getCookie(r *http.Request) string {

	fmt.Println("Full header = %v", r.Header)
	cookie, err := r.Cookie("chomp_sessionid")
	if err != nil {
		fmt.Println("Error..cookie = %v, err:%v, cookie1:%v err1:%v",cookie, err)
		return ""
	}
	fmt.Println("Cookie = %v", cookie)

	if cookiestr := r.Header.Get("Cookie"); cookiestr == "" {
		return ""
	} else {
		parts := strings.Split(strings.TrimSpace(cookiestr), ";")
		for k, v := range parts {
			nameval := strings.Split(v, "=")
			if k == 0 && nameval[0] != "chomp_sessionid" {
				return ""
			} else {
				fmt.Printf("Returning cookie %v\n", nameval[1])
				return nameval[1]
			}
		}
	}
	return ""
}