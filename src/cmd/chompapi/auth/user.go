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
)

const (
	TokenLength int           = 32
	TtlDuration time.Duration = 20 * time.Minute
)

type User struct {
	Id        int64       `db:"id"`
	Email     string      `db:"email"`
	Token     string      `db:"token"`
	Ttl       time.Time   `db:"ttl"`
	OriginUrl null.String `db:"originurl"`
}

// RefreshToken refreshes Ttl and Token for the User.
func (u *User) RefreshToken() error {
	token := make([]byte, TokenLength)
	if _, err := io.ReadFull(rand.Reader, token); err != nil {
		return err
	}
	u.Token = base64.URLEncoding.EncodeToString(token)
	u.Ttl = time.Now().UTC().Add(TtlDuration)
	return nil
}

// IsValidToken returns a bool indicating that the User's current token hasn't
// expired and that the provided token is valid.
func (u *User) IsValidToken(token string) bool {
	if u.Ttl.Before(time.Now().UTC()) {
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
						WHERE id=?`, user.Id).Scan(&user.Id, &user.Email, &user.Token,
					   	&user.OriginUrl)
	if err != nil {
		fmt.Printf("\nSQL err = %v\n", err)
		return err
	}
	return err
}

func (user *User) SetUserInfo(db *sql.DB) error {
	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", user)
	fmt.Printf("Type of userInfo = %v\n", reflect.TypeOf(user))

	// query := fmt.Sprintf("INSERT INTO chomp_users SET chomp_username='%s', email='%s', phone_number='%s', password_hash='%s', dob='%d', gender='%s'", 
	// 	userInfo.Username, userInfo.Email, userInfo.Phone, userInfo.Hash, userInfo.Dob, userInfo.Gender)
	// fmt.Println("Query = %v\n", query)
	// myUuid := uuid.NewRandom()
	// fmt.Printf("Udid = %v\n", myUuid.String())
	// user.Token = myUuid.String()

	results, err := db.Exec(`INSERT INTO signup_verification
							SET id=?, token=?, email=?, ttl=?, origin_url=?`, 
							user.Id, user.Token, user.Email, time.Now().Unix() + 1776600, "user.OriginUrl")

	if err != nil {
		fmt.Printf("Update Account Setup Time err = %v\n", err)
		return err
	}
	
	id, err := results.LastInsertId()
	user.Id = int64(id)
	fmt.Printf("Results = %v\n err3 = %v\n", user.Id , err)
	fmt.Printf("Error = %v\n", err)
	return nil
}
