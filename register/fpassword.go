package register

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"chompapi/db"
	"chompapi/crypto"
	"time"
	"chompapi/globalsessionkeeper"
	"chompapi/me"
)

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var myErrorResponse globalsessionkeeper.ErrorResponse

	switch r.Method {
	case "POST":
		input := new(db.UserInfo)
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			fmt.Printf("something %v", err)
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "Malformed JSON: " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
		}
		fmt.Printf("Json Input = %+v\n", input)
		fmt.Println("int = %v", input.Email)

		if isValidInput(input, &myErrorResponse) == false {
			fmt.Println("Something not valid")
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		randomPass = GeneratePassword(8)

		input.Hash = hex.EncodeToString(crypto.GeneratePassword(input.Username, []byte(randomPass)))
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
