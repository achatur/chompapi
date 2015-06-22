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

type UserInfo struct {
	ChompUserID   int
	ChompUsername string
	Email         string
	PhoneNumber   string
	PasswordHash  string
	DOB           string
	Gender        string
	PhotoID 	  int
}

// Plurals are names of tables in DB
// while the singular form of the structs
// are the inputs from json

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
type Photo struct {
	ID 	int
}

type Reviews struct {
	ID 				int
	DishID			int
	UserID 			int
	Username 		string
	RestaurantID	int
	PhotoID 		int
	Price			float64
	Like			bool
	Descr			string
	Complete		bool
}
type Review struct {
	id 				int
	Username 		string
	UserID 			int
	Restaurant 		Restaurants
	Dish 			Dish
	Photo 			Photo
	Price 			float32
	Liked 			bool
	Description 	string
}

type Dish struct {
	ID 				int
	Name 			string
}

type Restaurants struct {
	ID				int
	Name 			string
	Latt			float64
	Long			float64
	LocationNum		int
	Source			string
	SourceLocID		string
}
// type Restaurant struct {
//   	Name			string
//   	Latt			float64
//   	Long			float64
//   	Source			string
//   	SourceLocID 	string
// }

func (userInfo *UserInfo) GetUserInfo(username string) error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for reading chomp_users table data
	fmt.Printf("SELECT * FROM chomp_users WHERE chomp_username=%s\n", userInfo.ChompUsername)
	err = db.QueryRow("SELECT * FROM chomp_users WHERE chomp_username=?", 
					   userInfo.ChompUsername).Scan(&userInfo.ChompUserID, &userInfo.Email,
					   							    &userInfo.ChompUsername, &userInfo.PhoneNumber,
					   							    &userInfo.PasswordHash,&userInfo.DOB,
					   							    &userInfo.Gender, &userInfo.PhotoID)
	if err != nil {
		fmt.Printf("err = %v", err)
		return err
	}
	// columns, err := rows.Columns()
	// if err != nil {
	// 	return err
	// }
	// values := make([]sql.RawBytes, len(columns))
	// scanArgs := make([]interface{}, len(values))
	// for i := range values {
	// 	scanArgs[i] = &values[i]
	// }
	// fmt.Println("scanArgs = %v\n", scanArgs)
	// for rows.Next() {
	// 	err = rows.Scan(scanArgs...)
	// 	if err != nil {
	// 		return make(map[string]string), err
	// 	}
	// 	var value string
	// 	for i, col := range values {
	// 		if col == nil {
	// 			value = "null"
	// 		} else {
	// 			value = string(col)
	// 		}
	// 		m[columns[i]] = value
	// 		fmt.Println(columns[i], ": ", value)
	// 	}
	// 	fmt.Println("--------------------------------")
	// }
	// if err = rows.Err(); err != nil {
	// 	return make(map[string]string), err
	// }
	return err
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
		err = errors.New("Could noterrors return photo info")
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

func (review *Reviews) SetReview() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", review)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(review))

	_, err = db.Query("INSERT INTO reviews SET ", review.ID)

	return err
}

func (review *Reviews) UpdateReview() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", review)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(review))

	_, err = db.Query("DELETE FROM photos WHERE id=?", review.ID)

	return err
}

func (review *Reviews) DeleteReview() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", review)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(review))

	_, err = db.Query("DELETE FROM photos WHERE id=?", review.ID)

	return err
}

func (dish *Dish) GetDishInfoByName() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("inside call: restaurants = %v\n", dish)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(dish))
	fmt.Println("")

	err2 := db.QueryRow(`SELECT id, name
						FROM dish
						WHERE BINARY name=?`,dish.Name).Scan(&dish.ID, &dish.Name)

	fmt.Println("Inside DB: dish now: ", dish)
 	fmt.Println("Error: ", err2)

	return err2
}

func (dish *Dish) CreateDish() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("Creating dish = %v\n", dish)
	fmt.Printf("Type of dish = %v\n", reflect.TypeOf(dish))
	fmt.Println("")
	fmt.Printf("INSERT INTO dish SET name='%s'\n", dish.Name)

	// err2 := db.QueryRow(`INSERT INTO dish
	// 					 SET name=?`,dish.Name).Scan(&dish.ID, dish.Name)

	results, err2 := db.Exec(`INSERT INTO dish SET name=?`, dish.Name)

	id, err2 := results.LastInsertId()
	dish.ID = int(id)

	fmt.Printf("Results = %v\n err3 = %v\n", id , err2)

	return err2


	fmt.Println("Inside DB: dish now: ", dish)
 	fmt.Println("Error: ", err2)

	return err2
}

func (restaurant *Restaurants) GetRestaurantInfoByName() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("inside call: restaurants = %v\n", restaurant)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(restaurant))
	fmt.Println("")

	err2 := db.QueryRow(`SELECT id, name, latitude, longitude, location_num, source, source_location_id
						FROM restaurants
						WHERE name=?`,restaurant.Name).Scan(&restaurant.ID, &restaurant.Name,
															  &restaurant.Latt, &restaurant.Long,
															  &restaurant.LocationNum, &restaurant.Source,
															  &restaurant.SourceLocID)
	fmt.Println("Inside DB: restaurant now: ", restaurant)
 	fmt.Println("Error: ", err2)

	return err2
}

func (restaurant *Restaurants) CreateRestaurant() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("inside call: restaurants = %v\n", restaurant)
	fmt.Printf("Type of userInfo = %v\n\n", reflect.TypeOf(restaurant))

	results, err2 := db.Exec(`INSERT INTO restaurants
						 SET id = ?, name = ?, latitude = ?, longitude = ?,
						 location_num = ?, source = ?,
						 source_location_id = ?`, restaurant.ID, restaurant.Name,
						 					      restaurant.Latt, restaurant.Long,
						 	  					  restaurant.LocationNum, restaurant.Source,
						 	  					  restaurant.SourceLocID)
	id, err2 := results.LastInsertId()
	restaurant.ID = int(id)

	fmt.Printf("Results = %v\n err3 = %v\n", id , err2)

	return err2
}

func (review *Review) CreateReview() int {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return -1
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", review)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(review))

	// err2 := db.QueryRow(`INSERT into review SET (user_id = ?, username = ?, dish_id ? =, 
	// 											 photo_id =?, restaurant_id = ?, 
	// 											 price = ?, like = ?, complete = ?, description = ?)
	// 					FROM restaurants
	// 					WHERE name='?'`,restaurant.Name).Scan(&review.ID, &restaurant.ID, &restaurant.Name,
	// 														  &restaurant.Latt, &restaurant.Long,
	// 														  &restaurant.LocationNum, &restaurant.Source,
	// 														  &restaurant.SourceLocID)
	// if err2 != sql.ErrNoRows {
	// 	return 1
	// } else if err2 != nil {
	// 	return -1
	// }
	return 0
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
