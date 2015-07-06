package register

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"chompapi/db"
	"chompapi/crypto"
	"time"
	"chompapi/globalsessionkeeper"
	"math/rand"
	"chompapi/messenger"
)

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var myErrorResponse globalsessionkeeper.ErrorResponse

	switch r.Method {
	case "POST":

		input := new(db.UserInfo)
		dbUserInfo := new(db.UserInfo)
		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(&input); err != nil {
			fmt.Printf("something %v", err)
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "Malformed JSON: " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
		}

		fmt.Printf("Json Input = %+v\n", input)
		fmt.Println("int = %v", input.Email)

		if isValidInputUser(input, &myErrorResponse) == false {
			fmt.Println("Something not valid")
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		dbUserInfo.Email = input.Email

		if err := dbUserInfo.GetUserInfoByEmail(); err != nil {
			fmt.Printf("Could not find user")
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "User Not Found " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		if dbUserInfo.DOB != input.DOB {
			fmt.Printf("DOB does not match")
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Error = "DOB Does not Match"
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		randomPass := GeneratePassword(13)
		fmt.Printf("RandomPass = %v\n", randomPass)

		input.PasswordHash = hex.EncodeToString(crypto.GeneratePassword(dbUserInfo.Username, []byte(randomPass)))
		fmt.Printf("Hash = %s\n", input.PasswordHash)
		input.UserID = dbUserInfo.UserID

		if err := input.UpdatePassword(true); err != nil {
			fmt.Println("Error! = %v\n", err)
			myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = "Could not Update Password: " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		// Send email
		fmt.Println("Sending Email...")
		body := fmt.Sprintf("Your password has been reset, here's your nnew password\n\n%v\n\nRegards,\n\nThe Chomp Team", randomPass)
		context := new(messenger.SmtpTemplateData)
	    context.From = "Chomp"
	    context.To = input.Email
	    context.Subject = "Password Reset"
	    context.Body = body
	    context.Pass = randomPass

	    err := context.SendGmail()
	    if err != nil {
	    	fmt.Printf("Something ewnt wrong %v\n", err)
	    	myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Error = "Could not send mail" + err.Error()
			myErrorResponse.HttpErrorResponder(w)
	    }

	    fmt.Printf("Mail sent")
		w.WriteHeader(http.StatusNoContent)
		return
		
	default:

		myErrorResponse.Code = http.StatusMethodNotAllowed
		myErrorResponse.Error = "Invalid Method"
		myErrorResponse.HttpErrorResponder(w)
		return
	}
}

func isValidInputUser(userInfo *db.UserInfo, errorResponse *globalsessionkeeper.ErrorResponse) bool {
	if isValidString(userInfo.Email) == false {
		fmt.Println("not valid email = ", userInfo.Email)
		errorResponse.Error = "Invalid Email " + userInfo.Email
		return false
	} else if userInfo.DOB == 0 {
		errorResponse.Error = "Invalid DOB " + strconv.Itoa(userInfo.DOB)
		return false
	}
	return true
}

func GeneratePassword(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789~!@#$%^^&*()_+`-=")
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}