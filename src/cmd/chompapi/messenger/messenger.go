package messenger

import (
	"net/smtp"
	"fmt"
	// "text/template"
	"html/template"
	"bytes"
	"io/ioutil"
	// "strconv"
	"encoding/json"
)

// const emailTemplate = `From: {{.From}}
// To: {{.To}}
// Subject: {{.Subject}}
// MIME-version: 1.0
// Content-Type: text/html; charset="UTF-8"
// Content-Transfer-Encoding: quoted-printable

// Hello!
// {{if .Pass}}
// We have reset your password at your request:</br>
// </br>
// {{.Pass}}</br></br>
// {{else if .Username}}
// We recently recieved word that you forgot your username.  Here's your username:</br>
// </br>
// {{.Username}}</br></br>
// {{else}}
// This is just an email to inform you that you changed your password recently.</br>
// {{end}}
// Feel free to delete this email and carry on enjoying your food!</br>
// </br>
// <a href="chompapp://">Login To App</a></br>
// </br>
// All the best,</br>
// </br>
// The Chomp Team</br></body></html>` 

type EmailUser struct {
    Username    string
    Password    string
    EmailServer string
    Port        int
}

type SmtpTemplateData struct {
    From    string
    To      string
    Subject string
    Body    string
    Pass 	string
    Username string
}

func (smtpTemplateData *SmtpTemplateData) SendGmail() error {

	fmt.Printf("smtp data = %v\n", smtpTemplateData)
	emailUser := new(EmailUser)
	fileContent, err := ioutil.ReadFile("./chomp_private/email.json")
	// fileContent, err := ioutil.ReadFile("./chomp_private/email_mandrill.json")

	if err != nil {

		fmt.Printf("Could not open file")
		return err
	}

	emailTemplateByte, err := ioutil.ReadFile("./messenger/email_template_2.html")
	if err != nil {
	    return err
	}
	emailTemplate := string(emailTemplateByte[:])
	err = json.Unmarshal(fileContent, &emailUser)

	if  err != nil {

        fmt.Printf("Err = %v", err)
        return err
    }

	auth := smtp.PlainAuth(

		"",
    	emailUser.Username,
    	emailUser.Password,
    	emailUser.EmailServer)

	var doc bytes.Buffer
	t := template.New("emailTemplate")

	if t, err = t.Parse(emailTemplate); err != nil {
	    fmt.Print("error trying to parse mail template")
	    return err
	}


	if err = t.Execute(&doc, smtpTemplateData); err != nil {
	    fmt.Print("error trying to execute mail template")
	    return err
	}

	//sending mail
	err = smtp.SendMail(emailUser.EmailServer+":587", // in our case, "smtp.google.com:587"
    auth,
    emailUser.Username,
    []string{smtpTemplateData.To},
    // []string{"amir.chatur@gmail.com"},
    doc.Bytes())
	if err != nil {
    	fmt.Print("ERROR: attempting to send a mail ", err)
    	return err
	}
	return nil
}