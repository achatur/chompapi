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
	// "os"
	// "io"
)

// type EmailUser struct {
//     Username    string
//     Password    string
//     EmailServer string
//     Port        int
// }

// type SmtpTemplateData struct {
//     From    string
//     To      string
//     Subject string
//     Body    string
//     Pass 	string
//     Username string
//     Link	string
// }

func (smtpTemplateData *SmtpTemplateData) SendGmailVerify() error {

	fmt.Printf("smtp data = %v\n", smtpTemplateData)
	emailUser := new(EmailUser)
	fileContent, err := ioutil.ReadFile("./chomp_private/email.json")
	// fileContent, err := ioutil.ReadFile("./chomp_private/email_mandrill.json")

	if err != nil {

		fmt.Printf("Could not open file")
		return err
	}

	emailTemplateByte, err := ioutil.ReadFile("./messenger/email_template_verify_email.html")
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
	    fmt.Printf("error trying to parse mail template, %v\n", err)
	    return err
	}

	fmt.Printf("template = %v\n", emailTemplate)

	if err = t.Execute(&doc, smtpTemplateData); err != nil {
	    fmt.Print("error trying to execute mail template, %v\n", err)
	    return err
	}

	//sending mail
	fmt.Printf("Doc = %v\n", doc.String())
	//doc.WriteTo(os.Stdout)
	// os.Stdout.Write(doc)
	// io.Copy(os.Stdout, doc)

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