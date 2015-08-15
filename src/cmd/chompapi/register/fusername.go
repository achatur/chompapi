package register

import (
	"encoding/json"
	"fmt"
	"net/http"
	"cmd/chompapi/db"
	"cmd/chompapi/globalsessionkeeper"
	"cmd/chompapi/messenger"
)

func ForgotUsername(w http.ResponseWriter, r *http.Request) {
	var myErrorResponse globalsessionkeeper.ErrorResponse

	switch r.Method {
	case "POST":

		input := new(db.UserInfo)
		dbUserInfo := new(db.UserInfo)
		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(&input); err != nil {
			fmt.Printf("something %v", err)
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Desc= "Malformed JSON: " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
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
			myErrorResponse.Desc= "User Not Found " + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
		}
		fmt.Printf("DbUserUnfo = %v\n", dbUserInfo)
		if dbUserInfo.DOB != input.DOB {
			fmt.Printf("DOB does not match")
			myErrorResponse.Code = http.StatusBadRequest
			myErrorResponse.Desc= "DOB Does not Match"
			myErrorResponse.HttpErrorResponder(w)
			return
		}

		// Send email
		fmt.Println("Sending Email...")
		body := fmt.Sprintf("Your username is:\n\n%v\n\nRegards,\n\nThe Chomp Team", dbUserInfo.Username)
		context := new(messenger.SmtpTemplateData)
	    context.From = "Chomp"
	    context.To = input.Email
	    context.Subject = "Forgot Login Information"
	    context.Body = body
	    context.Username = dbUserInfo.Username

	    fmt.Printf("Context = %v\n", context)

	    err := context.SendGmail()
	    if err != nil {
	    	fmt.Printf("Something ewnt wrong %v\n", err)
	    	myErrorResponse.Code = http.StatusInternalServerError
			myErrorResponse.Desc= "Could not send mail" + err.Error()
			myErrorResponse.HttpErrorResponder(w)
			return
	    }

	    fmt.Printf("Mail sent")
		w.WriteHeader(http.StatusNoContent)
		return
		
	default:

		myErrorResponse.Code = http.StatusMethodNotAllowed
		myErrorResponse.Desc= "Invalid Method"
		myErrorResponse.HttpErrorResponder(w)
		return
	}
}
