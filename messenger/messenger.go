package messenger

import (
	"net/smtp"
	"fmt"
	"text/template"
	"bytes"
	"io/ioutil"
	"strconv"
	"encoding/json"
	// "chompapi/globalsessionkeeper"
	// "errors"
	// "net/http"
)

// const emailTemplate = `From: &#123;&#123;.From&#125;&#125;
// To: &#123;&#123;.To&#125;&#125;
// Subject: &#123;&#123;.Subject&#125;&#125;

// &#123;&#123;.Body&#125;&#125;

// Sincerely,

// &#123;&#123;.From&#125;&#125;
// `

const emailTemplate = `From: &amp;#123;&amp;#123;.From&amp;#125;&amp;#125;
To: &amp;#123;&amp;#123;.To&amp;#125;&amp;#125;
Subject: &amp;#123;&amp;#123;.Subject&amp;#125;&amp;#125;

Hello!

We have reset your password at your request:

%v

Feel free to delete this email and carry on enjoying your food!

BTW, we can add HTML and make this email a lot prettier.  This is just a POC.

All the best,

The Chomp Team` 

const emailTemplateNopass = `From: &amp;#123;&amp;#123;.From&amp;#125;&amp;#125;
To: &amp;#123;&amp;#123;.To&amp;#125;&amp;#125;
Subject: &amp;#123;&amp;#123;.Subject&amp;#125;&amp;#125;

Hello!

This is just an email to inform you that you changed your password recently.

Feel free to delete this email and carry on enjoying your food!

All the best,

The Chomp Team` 

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
}

func (smtpTemplateData *SmtpTemplateData) SendGmail() error {

	// var myErrorResponse globalsessionkeeper.ErrorResponse
	emailUser := new(EmailUser)
	// err := errors.New("")
	 fileContent, err := ioutil.ReadFile("./chomp_private/email.json")

	if err != nil {

		fmt.Printf("Could not open file")
		return err
	}
	
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
	var message string
	if smtpTemplateData.Pass != "" {
		message = fmt.Sprintf(emailTemplate, smtpTemplateData.Pass)
	} else {
		message = fmt.Sprintf(emailTemplateNopass)
	}
	
	t := template.New("emailTemplate")

	if t, err = t.Parse(message); err != nil {
	    fmt.Print("error trying to parse mail template")
	    return err
	}


	if err = t.Execute(&doc, smtpTemplateData); err != nil {
	    fmt.Print("error trying to execute mail template")
	    return err
	}

	//sending mail
	err = smtp.SendMail(emailUser.EmailServer+":"+strconv.Itoa(emailUser.Port), // in our case, "smtp.google.com:587"
    auth,
    emailUser.Username,
    //[]string{"amir.chatur@gmail.com"},
    []string{smtpTemplateData.To},
    doc.Bytes())
	if err != nil {
    	fmt.Print("ERROR: attempting to send a mail ", err)
    	return err
	}
	return nil

}