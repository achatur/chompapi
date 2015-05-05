package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/url"
	"reflect"
	"strconv"
)

func getUserInfo(username string) (map[string]string, err error) {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		//panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
		return (map[string]string, err)
	}
	defer db.Close()
	m := map[string]string{}

	// Prepare statement for reading chomp_users table data
	query := "SELECT * FROM chomp_users WHERE chomp_username="
	query += "'"
	query += username
	query += "'"
	//rows, err := db.Query(string(query))
	rows, err := db.Query("SELECT * FROM chomp_users WHERE chomp_username=?", username)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		return "", err 
	}
	columns, err := rows.Columns()
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		return "", err
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	fmt.Println("scanArgs = %v\n", scanArgs)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			//panic(err.Error())
			return "", err
		}
		var value string
		for i, col := range values {
			if col == nil {
				value = "null"
			} else {
				value = string(col)
			}
			m[columns[i]] = value
			fmt.Println(columns[i], ": ", value)
		}
		fmt.Println("--------------------------------")
	}
	if err = rows.Err(); err != nil {
		//panic(err.Error())
		return "", err
	}
	return m, err
}

func (userInfo RegisterInput) SetUserInfo() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		//panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", userInfo)
	fmt.Println("Type of userInfo = %w\n", reflect.TypeOf(userInfo))
	userMap := structToMap(&userInfo)
	fmt.Println("struct2map = %v\n", userMap)

	query := fmt.Sprintf("INSERT INTO chomp_users SET chomp_username='%s', email='%s', phone_number='%s', password_hash='%s', dob='%s', gender='%s'", userInfo.Username, userInfo.Email, userInfo.Phone, userInfo.Hash, userInfo.Dob, userInfo.Gender)
	fmt.Println("Query = %v\n", query)

	stmt, err := db.Prepare(query)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		return err
	}
	return nil
}

func isValid(s sql.NullString) string {

	if s.Valid {
		fmt.Println("s is valid")
		return s.String
	} else {
		fmt.Println("s is not valid")
		return s.String
	}
}

func structToMap(i interface{}) (values url.Values) {
	values = url.Values{}
	iVal := reflect.ValueOf(i).Elem()
	typ := iVal.Type()
	for i := 0; i < iVal.NumField(); i++ {
		f := iVal.Field(i)
		// You ca use tags here...
		// tag := typ.Field(i).Tag.Get("tagname")
		// Convert each type into a string for the url.Values string map
		var v string
		switch f.Interface().(type) {
		case int, int8, int16, int32, int64:
			v = strconv.FormatInt(f.Int(), 10)
		case uint, uint8, uint16, uint32, uint64:
			v = strconv.FormatUint(f.Uint(), 10)
		case float32:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 32)
		case float64:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 64)
		case []byte:
			v = string(f.Bytes())
		case string:
			v = f.String()
		}
		values.Set(typ.Field(i).Name, v)
	}
	return
}
