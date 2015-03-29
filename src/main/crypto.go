package main

import (
	"bytes"
	//"crypto/rand"
	"crypto/sha1"
	"fmt"
	//"hash"
	"io"
	//"os"
)

const saltSize = 0

func generateSalt(secret []byte) []byte {
	buf := make([]byte, saltSize, saltSize+sha1.Size)
	//_, err := io.ReadFull(rand.Reader, buf)

	//if err != nil {
	//	fmt.Printf("random read failed: %v", err)
	//	os.Exit(1)
	//}

	hash := sha1.New()
	hash.Write(buf)
	hash.Write(secret)
	return hash.Sum(buf)
}

func generatePassword(username string, password []byte) []byte {

	fmt.Printf("Password : %s\n", string(password))
	salt := generateSalt([]byte(username))
	fmt.Printf("salt : %x\n", string(salt))

	//generate password + salt to store into db
	combination := string(salt) + string(password)
	passwordHash := sha1.New()
	io.WriteString(passwordHash, combination)
	fmt.Printf("password Hash : %x \n", passwordHash.Sum(nil))
	return (passwordHash.Sum(nil))
}

func validatePassword(username string, password []byte, dbPasswordHash string) bool {

	passwordHash := generatePassword(username, password)
	//temp := hash.Hash([]byte(dbPasswordHash))
	fmt.Printf("Validate hash gen = %v\n", passwordHash)

	match := bytes.Equal(passwordHash, []byte(dbPasswordHash))
	return match
}
