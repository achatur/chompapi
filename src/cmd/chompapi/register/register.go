package register

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"cmd/chompapi/db"
	"cmd/chompapi/crypto"
	"time"
	"cmd/chompapi/globalsessionkeeper"
	"cmd/chompapi/me"
)

func DoRegister(w http.ResponseWriter, r *http.Request) {
	var myErrorResponse globalsessionkeeper.ErrorResponse

	switch r.Method {
	case "POST":
		// input := newUser()
		input := new(db.RegisterInput)
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			fmt.Printf("something %v", err)
		}
		fmt.Printf("Json Input = %+v\n", input)
		fmt.Println("int = %v", input.Dob)

		if isValidInput(input, &myErrorResponse) == false {
			fmt.Println("Something not valid")
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.HttpErrorResponder(w)

			return
		}

		input.Hash = hex.EncodeToString(crypto.GeneratePassword(input.Username, []byte(input.Password)))
		fmt.Printf("Hash = %s\n", input.Hash)

		err := input.SetUserInfo()
		if err != nil {
			fmt.Println("Error! = %v\n", err)
			if strings.Contains(err.Error(), "Error 1062") {
				myErrorResponse.Code = http.StatusConflict
				myErrorResponse.Error = "Duplicate Not Allowed:-:" + err.Error()
				myErrorResponse.HttpErrorResponder(w)
				return
			}

			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		// set user photo now
		var photoInfo db.Photos
		photoInfo.Uuid = me.GenerateUuid()
		photoInfo.Username = input.Username
		// *photoInfo.Latitude = 0.0
		// *photoInfo.Longitude = 0.0
		err = photoInfo.SetMePhoto()
			if err != nil {
				//need logging here instead of print
				myErrorResponse.Code = http.StatusInternalServerError
				myErrorResponse.Error = err.Error()
				myErrorResponse.HttpErrorResponder(w)
				return
			} 

		err = photoInfo.GetPhotoInfoByUuid()
		if err != nil {
			//need logging here instead of print
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		input.Photo.ID = photoInfo.ID
		err = photoInfo.UpdatePhotoIDUserTable()
		if err != nil {
			//need logging here instead of print
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		igStore 	 := new(db.IgStore)
		igStore.UserID = input.UserID
		igStore.IgMediaID = "fake"
		igStore.IgCreatedTime, err = strconv.Atoi(instaData.Data[i].CreatedTime)
		err = igStore.UpdateLastPull()
		if err != nil {
			fmt.Printf("Could not update table\n")
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = "IG UpdateLastPull failed: " + err.Error()
			return
		}

		w.Header().Set("Location", fmt.Sprintf("https://chompapi.com/me/photos/%v",  photoInfo.ID))
		w.Header().Set("UUID", photoInfo.Uuid)
		w.WriteHeader(http.StatusNoContent)
		return

	default:

		myErrorResponse.Code = http.StatusMethodNotAllowed
		myErrorResponse.Error = "Invalid Method"
		myErrorResponse.HttpErrorResponder(w)
		return
	}
}

func isValidInput(userInfo *db.RegisterInput, errorResponse *globalsessionkeeper.ErrorResponse) bool {
	if isValidString(userInfo.Email) == false {
		fmt.Println("not valid email = ", userInfo.Email)
		errorResponse.Error = "Invalid Email " + userInfo.Email
		return false
	}
	if isValidString(userInfo.Username) == false {
		fmt.Println("not valid username", userInfo.Username)
		errorResponse.Error = "Invalid Username " + userInfo.Username
		return false
	}
	if isValidString(userInfo.Password) == false {
		fmt.Println("not valid password", userInfo.Password)
		errorResponse.Error = "Invalid Password " + userInfo.Password
		return false
	}
	if userInfo.Dob == 0 || age(time.Unix(int64(userInfo.Dob), 0)) < 18 {
		errorResponse.Error = "Invalid Age " + string(age(time.Unix(int64(userInfo.Dob), 0)))
		return false
	}
	
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

// func newUser() *db.RegisterInput {
// 	return &db.RegisterInput{}
// }

func age(birthday time.Time) int {
	fmt.Println("made it here")
	now := time.Now()
	years := now.Year() - birthday.Year()
	if now.YearDay() < birthday.YearDay(){
		years--
	}
	fmt.Println("Age = %v", years)
	return years
}
