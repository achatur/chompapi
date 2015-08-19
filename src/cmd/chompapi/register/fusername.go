package register

import (
	"encoding/json"
	"fmt"
	"net/http"
	"cmd/chompapi/db"
	"cmd/chompapi/globalsessionkeeper"
	"cmd/chompapi/messenger"
)

func ForgotUsername(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {
	var myErrorResponse globalsessionkeeper.ErrorResponse

	switch r.Method {
	case "POST":

		input := new(db.UserInfo)
		dbUserInfo := new(db.UserInfo)
		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(&input); err != nil {
			fmt.Printf("something %v", err)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Malformed JSON: " + err.Error()}
		}

		fmt.Printf("Json Input = %+v\n", input)
		fmt.Println("int = %v", input.Email)

		if isValidInputUser(input, &myErrorResponse) == false {
			fmt.Println("Something not valid")
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Malformed JSON"}
		}

		dbUserInfo.Email = input.Email

		if err := dbUserInfo.GetUserInfoByEmail(a.DB); err != nil {
			fmt.Printf("Could not find user")
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "User Not Found " + err.Error()}
		}
		fmt.Printf("DbUserUnfo = %v\n", dbUserInfo)
		if dbUserInfo.DOB != input.DOB {
			fmt.Printf("DOB does not match")
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "DOB Does not Match"}
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
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Could not send mail" + err.Error()}
	    }

	    fmt.Printf("Mail sent")
		w.WriteHeader(http.StatusNoContent)
		return nil
		
	default:

		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Invalid Method"}
	}
}
