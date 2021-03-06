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
	"cmd/chompapi/auth"
	"cmd/chompapi/messenger"
)

func DoRegister(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {
	var myErrorResponse globalsessionkeeper.ErrorResponse

	switch r.Method {
	case "POST":

		input := new(db.RegisterInput)
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil {
			fmt.Printf("something %v", err)
		}
		fmt.Printf("Json Input = %+v\n", input)
		fmt.Println("int = %v", input.Dob)

		if isValidInput(input, &myErrorResponse) == false {
			fmt.Println("Something not valid")
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, myErrorResponse.Desc}
		}

		input.Hash = hex.EncodeToString(crypto.GeneratePassword(input.Username, []byte(input.Password)))
		fmt.Printf("Hash = %s\n", input.Hash)

		err := input.SetUserInfo(a.DB)
		if err != nil {
			fmt.Println("Error! = %v\n", err)
			if strings.Contains(err.Error(), "Error 1062") {
				return globalsessionkeeper.ErrorResponse{http.StatusConflict, "Duplicate Not Allowed:-:" + err.Error()}
			}

			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
		}
		// set user photo now
		var photoInfo db.Photos
		photoInfo.Uuid = me.GenerateUuid()
		photoInfo.Username = input.Username

		err = photoInfo.SetMePhoto(a.DB)
			if err != nil {
				//need logging here instead of print
				return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
			} 

		err = photoInfo.GetPhotoInfoByUuid(a.DB)
		if err != nil {
			//need logging here instead of print
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
		}
		input.Photo.ID = photoInfo.ID
		err = photoInfo.UpdatePhotoIDUserTable(a.DB)
		if err != nil {
			//need logging here instead of print
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
		}
		igStore := new(db.IgStore)
		igStore.UserID = input.UserID
		igStore.IgMediaID = "fake"
		igStore.IgCreatedTime = int(time.Now().Unix())
		err = igStore.UpdateLastPull(a.DB)
		if err != nil {
			fmt.Printf("Could not update table\n")
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "IG UpdateLastPull failed: " + err.Error()}
		}
		// Got to come back and fix this.  Database failures and email failures are causing
		// API to fail
		verifyUser := new(auth.User)
		fmt.Printf("Int here = %v\n", input.UserID)
		fmt.Printf("Int formatted here = %v\n", int64(input.UserID))
		verifyUser.Id = int64(input.UserID)
		verifyUser.Token = me.GenerateUuid()
		verifyUser.Email = input.Email
		fmt.Printf("Full map = %v\n", verifyUser)
		// err = verifyUser.SetUserInfo(a.DB)
		err = verifyUser.SetOrUpdateEmailVerify(a.DB)
		if err != nil {
			fmt.Printf("Could not add Verify User Info\n")
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "Could not add to verify table: " + err.Error()}
		}
		fmt.Println("Sending Email...")
		body := fmt.Sprintf("Your password was recently changed.\n\nRegards,\n\nThe Chomp Team")
		context := new(messenger.SmtpTemplateData)
	    context.From = "The Chomp Team"
	    context.To = input.Email
	    context.Subject = "Verify Email"
	    context.Body = body

	    err = context.SendGmail()
	    if err != nil {
	    	fmt.Printf("Can't Send Mail:\n&&&&&&&&\n%v\n&&&&&&&&\n", err)
			// return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "Could not send mail" + err.Error()}
	    }

	    fmt.Printf("Mail sent")
		w.WriteHeader(http.StatusNoContent)
		w.Header().Set("Location", fmt.Sprintf("https://chompapi.com/me/photos/%v",  photoInfo.ID))
		w.Header().Set("UUID", photoInfo.Uuid)
		w.WriteHeader(http.StatusNoContent)
		return nil

	default:
		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Invalid Method"}
	}
}

func isValidInput(userInfo *db.RegisterInput, errorResponse *globalsessionkeeper.ErrorResponse) bool {
	if isValidString(userInfo.Email) == false {
		fmt.Println("not valid email = ", userInfo.Email)
		errorResponse.Desc = "Invalid Email " + userInfo.Email
		return false
	}
	if isValidString(userInfo.Username) == false {
		fmt.Println("not valid username", userInfo.Username)
		errorResponse.Desc = "Invalid Username " + userInfo.Username
		return false
	}
	if isValidString(userInfo.Password) == false {
		fmt.Println("not valid password", userInfo.Password)
		errorResponse.Desc = "Invalid Password " + userInfo.Password
		return false
	}
	if userInfo.Dob == 0 || age(time.Unix(int64(userInfo.Dob), 0)) < 18 {
		errorResponse.Desc = "Invalid Age " + string(age(time.Unix(int64(userInfo.Dob), 0)))
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
