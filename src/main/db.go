package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func getUserInfo(username string) map[string]string {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()
	m := map[string]string{}

	// Prepare statement for reading chomp_users table data
	query := "SELECT * from chomp_users where chomp_username="
	query += "'"
	query += username
	query += "'"
	rows, err := db.Query(string(query))
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
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
			panic(err.Error())
		}
		var value string
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			m[columns[i]] = value
			fmt.Println(columns[i], ": ", value)
		}
		fmt.Println("--------------------------------")
	}
	if err = rows.Err(); err != nil {
		panic(err.Error())
	}
	return m
}
