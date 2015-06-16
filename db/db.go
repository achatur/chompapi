package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"errors"
	//"github.com/pborman/uuid"
	 //"gopkg.in/gorp.v1"
)

type RegisterInput struct {
	Username	 string
	Email   	 string
	Password	 string
	Dob     	 int
	Gender  	 string
	Fname    	 string
	Lname    	 string
	Phone     	 string
	Hash     	 string
	Photo		 Photo
}

type Photo struct {
	ID 	int
}

type Photos struct {
	ID			int
	DishID		int
	UserID		int
	FilePath	string
	FileHash	string
	TimeStamp	string
	Uuid		string
	Username 	string
}

func GetUserInfo(username string) (map[string]string, error) {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return make(map[string]string), err
	}
	defer db.Close()
	m := map[string]string{}

	// Prepare statement for reading chomp_users table data
	rows, err := db.Query("SELECT * FROM chomp_users WHERE chomp_username=?", username)
	if err != nil {
		return make(map[string]string), err
	}
	columns, err := rows.Columns()
	if err != nil {
		return make(map[string]string), err
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
			return make(map[string]string), err
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
		return make(map[string]string), err
	}
	return m, err
}

func GetMeInfo(username string) (map[string]string, error) {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return make(map[string]string), err
	}
	defer db.Close()
	m := map[string]string{}

	// Prepare statement for reading chomp_users table data
	rows, err := db.Query("select chomp_user_id, chomp_username, email, dob, gender, photo_id from chomp_users where chomp_username=?", username)
	if err != nil {
		return make(map[string]string), err
	}
	columns, err := rows.Columns()
	if err != nil {
		return make(map[string]string), err
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
			return make(map[string]string), err
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
		return make(map[string]string), err
	}
	return m, err
}


func (userInfo RegisterInput) SetUserInfo() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", userInfo)
	fmt.Println("Type of userInfo = %w\n", reflect.TypeOf(userInfo))

	query := fmt.Sprintf("INSERT INTO chomp_users SET chomp_username='%s', email='%s', phone_number='%s', password_hash='%s', dob='%d', gender='%s'", 
		userInfo.Username, userInfo.Email, userInfo.Phone, userInfo.Hash, userInfo.Dob, userInfo.Gender)
	fmt.Println("Query = %v\n", query)

	stmt, err := db.Prepare(query)
	if err != nil {
		fmt.Println("Error occurd")
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return nil
}

// func (userInfo RegisterInput) UpdateUserPhoto() error {
// 	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
// 	if err != nil {
// 		return err
// 	}
// 	defer db.Close()
// 	_, err = db.Query("UPDATE chomp_users set photo_id=? WHERE chomp_username=?", 
// 					userInfo.Photo.ID, userInfo.Username)

// 	return err
// }

// func (photos *PhotoTable) SetPhoto() error {
// 	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp", "parseTime=true")
// 	if err != nil {
// 		return err
// 	}
// 	defer db.Close()

// 	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
// 	err := dbmap.Insert(photos)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func (photo Photos) SetMePhoto() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", photo)
	fmt.Println("Type of userInfo = %w\n", reflect.TypeOf(photo))

	//query := fmt.Sprintf("INSERT INTO photos SET dish_id='%d', chomp_user_id='%d', file_path='%s', file_hash='%s', uuid='%s'", photo.DishID, photo.UserID, photo.FilePath, photo.FileHash, photo.Uuid)
	query := fmt.Sprintf("INSERT into photos(chomp_user_id, file_path, file_hash, uuid) SELECT chomp_user_id, '%s', '%s', '%s' from chomp_users WHERE chomp_username='%s'", 
						photo.FilePath, photo.FileHash, photo.Uuid, photo.Username)
	fmt.Println("Query = %v\n", query)

	stmt, err := db.Prepare(query)
	if err != nil {
		fmt.Println("Error occurd")
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return nil
}

func (photo Photos) UpdatePhotoIDUserTable() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", photo)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(photo))

	//query := fmt.Sprintf("INSERT INTO photos SET dish_id='%d', chomp_user_id='%d', file_path='%s', file_hash='%s', uuid='%s'", photo.DishID, photo.UserID, photo.FilePath, photo.FileHash, photo.Uuid)
	query := fmt.Sprintf("UPDATE chomp_users SET photo_id='%d' WHERE chomp_username='%s'", 
						photo.ID, photo.Username)
	fmt.Println("Query = %v\n", query)

	stmt, err := db.Prepare(query)
	if err != nil {
		fmt.Println("Error occurd")
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return nil
}


func (photo *Photos) GetPhotoInfoByUuid() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()
	// m := map[string]string{}

	// Prepare statement for reading chomp_users table data
	row := db.QueryRow("SELECT id, chomp_user_id, file_path, file_hash, time_stamp, uuid from photos where uuid=?", photo.Uuid).Scan(&photo.ID, &photo.UserID, &photo.FilePath, &photo.FileHash, &photo.TimeStamp, &photo.Uuid)
	fmt.Println("Row =", row)
	fmt.Println("Row Type = ", reflect.TypeOf(photo))
	if row != nil {
		err = errors.New("Could not return photo info")
	}
	return err
}

func (photo *Photos) GetMePhotoByUsername() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", photo)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(photo))

	err = db.QueryRow(`SELECT id, chomp_users.chomp_user_id, file_path, file_hash, time_stamp, uuid
						FROM photos
						JOIN chomp_users on photos.id = chomp_users.photo_id
						WHERE chomp_users.chomp_username=?`,photo.Username).Scan(&photo.ID, &photo.UserID, &photo.FilePath, &photo.FileHash, &photo.TimeStamp, &photo.Uuid)
	return err
}

func (photo *Photos) GetMePhotoByPhotoID() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", photo)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(photo))

	err = db.QueryRow(`SELECT chomp_user_id, file_path, file_hash, time_stamp, uuid
						FROM photos
						WHERE id=?`,photo.ID).Scan(&photo.UserID, &photo.FilePath, &photo.FileHash, &photo.TimeStamp, &photo.Uuid)
	return err
}


func (photo *Photos) UpdateMePhoto() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", photo)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(photo))

	_, err = db.Query("UPDATE photos set uuid=? WHERE id=?", 
					photo.Uuid, photo.ID)

	return err
}

func (photo *Photos) DeleteMePhoto() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", photo)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(photo))

	_, err = db.Query("DELETE FROM photos WHERE id=?", photo.ID)

	return err
}


func IsValid(s sql.NullString) string {

	if s.Valid {
		fmt.Println("s is valid")
		return s.String
	} else {
		fmt.Println("s is not valid")
		return s.String
	}
}
