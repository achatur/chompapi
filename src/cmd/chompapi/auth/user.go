package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"github.com/guregu/null"
	"io"
	"time"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	// "github.com/pborman/uuid"
	"cmd/chompapi/messenger"
	"net/url"
	"strconv"
)

const (
	TokenLength int           = 32
	TtlDuration time.Duration = 20 * time.Minute
)

type User struct {
	Id        int64       `db:"id"`
	Email     string      `db:"email"`
	Token     string      `db:"token"`
	// Ttl       time.Time   `db:"ttl"`
	Ttl       int64		   `db:"ttl"`
	OriginUrl null.String `db:"originurl"`
}

// RefreshToken refreshes Ttl and Token for the User.
func (u *User) RefreshToken() error {
	token := make([]byte, TokenLength)
	if _, err := io.ReadFull(rand.Reader, token); err != nil {
		return err
	}
	u.Token = base64.URLEncoding.EncodeToString(token)
	// u.Ttl = time.Now().UTC().Add(TtlDuration)
	u.Ttl = time.Now().Unix() + 1776600
	return nil
}

// IsValidToken returns a bool indicating that the User's current token hasn't
// expired and that the provided token is valid.
func (u *User) IsValidToken(token string) bool {
	// if u.Ttl.Before(time.Now().UTC()) {
	fmt.Printf("TTL = %v\n", u.Ttl)
	fmt.Printf("time now = %v\n", time.Now().Unix())
	if time.Now().Unix() >= u.Ttl {
		fmt.Println("Token Expired")
		return false
	}
	return subtle.ConstantTimeCompare([]byte(u.Token), []byte(token)) == 1
}

func (user *User) GetUserInfo(db *sql.DB) error {
	// Prepare statement for reading chomp_users table data
	fmt.Printf(`SELECT id, email, token, ttl, original_url
				FROM signup_verification
				WHERE id=%s`, user.Id, "\n")
	err := db.QueryRow(`SELECT id, email, token, ttl, origin_url
						FROM signup_verification
						WHERE id=?`, user.Id).Scan(&user.Id, &user.Email, &user.Token, &user.Ttl, 
					   	&user.OriginUrl)
	if err != nil {
		fmt.Printf("\nSQL err = %v\n", err)
		return err
	}
	results, err := db.Exec(`UPDATE signup_verification
							SET ttl=?
							WHERE id=?`, time.Now().Unix() - 17766001776600, user.Id)
	if err != nil {
		fmt.Printf("Update ttl getUserInfo err = %v\n", err)
		return err
	}
	id, err := results.LastInsertId()
	// user.Id = int64(id)
	fmt.Printf("Results = %v\n err3 = %v\n", id , err)
	fmt.Printf("Error = %v\n", err)
	return err
}

func (user *User) UpdateVerifed(db *sql.DB) error {
	// Prepare statement for reading chomp_users table data
	fmt.Printf(`UPDATE signup_verification
				SET verified=1, ttl=?
				WHERE id=%s`, time.Now().Unix() - 1776600, user.Id, "\n")
	results, err := db.Exec(`UPDATE signup_verification
							SET verified=1, ttl=?
							WHERE id=?`, time.Now().Unix() - 1776600, user.Id)
	if err != nil {
		fmt.Printf("Update  verified err = %v\n", err)
		return err
	}
	
	id, err := results.LastInsertId()
	user.Id = int64(id)
	fmt.Printf("Results = %v\n err3 = %v\n", user.Id , err)
	fmt.Printf("Error = %v\n", err)
	return nil
}

func (user *User) SetOrUpdateEmailVerify(db *sql.DB) error {
	// verifyUser := new(auth.User)
	// verifyUser.Id = int64(input.UserID)
	// verifyUser.Token = me.GenerateUuid()
	// verifyUser.Email = input.Email
	err := user.SetUserInfo(db)
	if err != nil {
		fmt.Printf("Could not add Verify User Info\n")
		// return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "Could not add to verify table: " + err.Error()}
		return err
	}
	fmt.Println("Sending Email...")
	fmt.Printf("User in here = %v\n", user)
	body := fmt.Sprintf("Please Verify Your Password.\n\nRegards,\n\nThe Chomp Team")
	context := new(messenger.SmtpTemplateData)
	context.From = "The Chomp Team"
	context.To = user.Email
	context.Subject = "Please Verify Your Email"
	context.Body = body
	// Build login url
	params := url.Values{}
	params.Add("token", user.Token)
	fmt.Printf("Int = %v\nformatted = %v\n", user.Id, strconv.FormatInt(user.Id, 10))
	params.Add("uid", strconv.FormatInt(user.Id, 10))
	// params.Add("uid", user.Uid)

	verifyUrl := url.URL{}

	// if r.URL.IsAbs() {
	// 	verifyUrl.Scheme = r.URL.Scheme
	// 	verifyUrl.Host = r.URL.Host
	// } else {
	verifyUrl.Scheme = "https"
	verifyUrl.Host = "api.chompapp.com"
	// }

	verifyUrl.Path = "/verify"

	// Send login email
	// var mailContent bytes.Buffer
	// ctx := struct {
	// 	verifyUrl string
	// }{
	// 	fmt.Sprintf("%s?%s", verifyUrl.String(), params.Encode()),
	// }
	context.Link = fmt.Sprintf("%s?%s", verifyUrl.String(), params.Encode())

	fmt.Printf("Context = %v\n", context)
	err = context.SendGmailVerify()
	if err != nil {
		fmt.Printf("Something ewnt wrong %v\n", err)
		// return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "Could not send mail" + err.Error()}
		return err
	}

	fmt.Printf("Mail sent")
	return nil
}

func (user *User) SetUserInfo(db *sql.DB) error {
	// Prepare statement for writing chomp_users table data
	fmt.Printf("map = %v\n", user)
	fmt.Printf("Type of userInfo = %v\n", reflect.TypeOf(user))

	results, err := db.Exec(`INSERT INTO signup_verification
							SET id=?, token=?, email=?, ttl=?, origin_url=?`, 
							user.Id, user.Token, user.Email, time.Now().Unix() + 1776600, "user.OriginUrl")

	if err != nil {
		fmt.Printf("Update Account Setup Time err = %v\n", err)
		return err
	}
	
	id, err := results.LastInsertId()
	fmt.Printf("Results = %v\n, ID = %v\nerr3 = %v\n", user.Id, id, err)
	fmt.Printf("Error = %v\n", err)
	return nil
}
