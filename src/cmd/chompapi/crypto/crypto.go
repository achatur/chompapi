package crypto

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"encoding/hex"
)

const saltSize = 0

func generateSalt(secret []byte) []byte {
	buf := make([]byte, saltSize, saltSize+sha1.Size)
	hash := sha1.New()
	hash.Write(buf)
	hash.Write(secret)
	return hash.Sum(buf)
}

func GeneratePassword(username string, password []byte) []byte {

	fmt.Printf("Password : %s\n", string(password))
	salt := generateSalt([]byte(username))
	fmt.Printf("salt : %x\n", string(salt))
	
	combination := string(salt) + string(password)
	passwordHash := sha1.New()
	io.WriteString(passwordHash, combination)
	fmt.Printf("password Hash : %x \n", passwordHash.Sum(nil))
	return (passwordHash.Sum(nil))
}

func ValidatePassword(username string, password []byte, dbPasswordHash string) bool {

	passwordHash := GeneratePassword(username, password)
	fmt.Printf("Validate hash gen = %x\n", passwordHash)
	decodedHexString, err := hex.DecodeString(dbPasswordHash)
	if err != nil {
		fmt.Println("Error = %v", err.Error())
	}
	fmt.Printf("dbPasswordHash = %x\n", decodedHexString)

	match := bytes.Equal(passwordHash, []byte(decodedHexString))
	return match
}
