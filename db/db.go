package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"errors"
	"time"
	"github.com/astaxie/beego/session"
	"chompapi/globalsessionkeeper"
	"strings"
	"strconv"
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
	UserID   		int 			`json:"userId"`
	Username 		string 			`json:"username"`
	Email         	string 			`json:"email"`
	PhoneNumber   	string 			`json:"phoneNumber,omitempty"`
	PasswordHash  	string 			`json:"passwordHash,omitempty"`
	Password 	  	string 			`json:"password,omitempty"`
	DOB           	int 			`json:"dob"`
	Gender        	string 			`json:"gender"`
	Photo 	  	  	Photo 			`json:"photo"`
	Fname 			string 		 	`json:"fname"`
	Lname 			string 		 	`json:"lname"`
	IsPasswordTemp 	bool 			`josn:"isPasswordTemp"`
	PasswordExpiry 	int 			`josn:"passwordExpiry"`
	InstaCode 		string 			`json:"instaCode,omitempty"`
}

// Plurals are names of tables in DB
// while the singular form of the structs
// are the inputs from json

type Photos struct {
	ID			int				`json:"id"`
	DishID		int				`json:"dishId"`
	UserID		int				`json:"userId"`
	FilePath	string			`json:"filePath"`
	FileHash	string			`json:"fileHash"`
	TimeStamp	int				`json:"timeStamp"`
	Uuid		string			`json:"uuid"`
	Username 	string			`json:"username"`
}
type Photo struct {
	ID 	int 	`json:"id"`
}

type Reviews struct {
	ID 				int
	DishID			int
	UserID 			int
	Username 		string
	RestaurantID	int
	PhotoID 		int
	Price			float64
	Liked			bool
	Descr			string
	Complete		bool
}

type Review struct {
	ID 				int 			`json:"id"`
	Username 		string			`json:"username"`
	UserID 			int				`json:"userId"`
	Restaurant 		Restaurants		`json:"restaurant"`
	Dish 			Dish			`json:"dish"`
	Photo 			Photo			`json:"photo"`
	Price 			float32			`json:"price"`
	Liked 			sql.NullBool	`json:"liked,omitempty"`
	Description 	string			`json:"description"`
	Finished		sql.NullBool	`json:"finished,omitempty"`
	DishTags		[]DishTag 		`json:"dishTags"`
	CreatedDate		int 			`json:"createdDate,omitempty"`
	LastUpdated 	int 			`json:"lastUpdated,omitempty"`
	FinishedTime 	*int 			`json:"finishedTime,omitempty"`
	Source 			string 			`json:"source"`
}

type DishTag struct {
	ID 				int 			`json:"id"`
	Tag 			string 			`json:"dishTag"`
}

type Crawl struct {
	Username 		string			`json:"username"`
	UserID 			int				`json:"userId"`
	InstaId 		string 			`json:"instaId"`
	InstaTok 		string 			`json:"instaTok"`
	Tags 			[]string 		`json:"tags"`
}


type Dish struct {
	ID 				int				`json:"id"`
	Name 			string			`json:"name"`
}

type Restaurants struct {
	ID				int				`json:"id"`
	Name 			string			`json:"name"`
	Latt			float64			`json:"latt"`
	Long			float64			`json:"long"`
	LocationNum		int				`json:"locationNum"`
	Source			string			`json:"source"`
	SourceLocID		string			`json:"sourceLocID"`
}

type IgStore struct {
	UserID 			int
	IgMediaID 		string
	Epoch 			int
	LastUpdated 	int
	IgCreatedTime 	int
}

func (userInfo *UserInfo) GetUserInfo() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for reading chomp_users table data
	fmt.Printf("SELECT * FROM chomp_users WHERE chomp_username=%s\n", userInfo.Username)
	err = db.QueryRow(`SELECT chomp_user_id, email, chomp_username,
						phone_number, password_hash, dob, gender, photo_id,
						is_password_temp, password_expiry, fname, lname, insta_code
					   FROM chomp_users
					   WHERE chomp_username=?`, 
					   userInfo.Username).Scan(&userInfo.UserID, &userInfo.Email,
					   							    &userInfo.Username, &userInfo.PhoneNumber,
					   							    &userInfo.PasswordHash,&userInfo.DOB,
					   							    &userInfo.Gender, &userInfo.Photo.ID,
					   							    &userInfo.IsPasswordTemp, &userInfo.PasswordExpiry,
					   							    &userInfo.Fname, &userInfo.Lname, &userInfo.InstaCode)
	if err != nil {
		fmt.Printf("err = %v", err)
		return err
	}
	return err
}

func (userInfo *UserInfo) GetUserInfoByEmail() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for reading chomp_users table data
	fmt.Printf("SELECT chomp_user_id, email, chomp_username,phone_number, password_hash, dob, gender, photo_id,fname = ?, lname = ? FROM chomp_users WHERE email=%s\n", userInfo.Email)

	err = db.QueryRow(`SELECT chomp_user_id, email, chomp_username,
						dob, gender, photo_id, is_password_temp, password_expiry,
						fname, lname
					   FROM chomp_users
					   WHERE email= ?`, 
					   userInfo.Email).Scan(&userInfo.UserID, &userInfo.Email,
					   							    &userInfo.Username, &userInfo.DOB,
					   							    &userInfo.Gender, &userInfo.Photo.ID,
					   							    &userInfo.IsPasswordTemp, &userInfo.PasswordExpiry,
					   							    &userInfo.Fname, &userInfo.Lname)
	if err != nil {
		fmt.Printf("err = %v\n", err)
		return err
	}
	return err
}

func (userInfo RegisterInput) SetUserInfo() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", userInfo)
	fmt.Printf("Type of userInfo = %v\n", reflect.TypeOf(userInfo))

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

func (userInfo *UserInfo) DeleteUser() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", userInfo)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(userInfo))

	_, err = db.Query("DELETE FROM chomp_users WHERE chomp_user_id=?", userInfo.UserID)

	fmt.Printf("Error = %v\n", err)

	return err
}

func (userInfo *UserInfo) SetUserInactive() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", userInfo)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(userInfo))

	_, err = db.Query(`UPDATE chomp_users SET active = ?
						WHERE chomp_user_id = ?`, false, userInfo.UserID)

	return err
}

func (userInfo UserInfo) UpdatePassword(temp bool) error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", userInfo)
	fmt.Printf("Type of userInfo = %v\n", reflect.TypeOf(userInfo))
	var results sql.Result
	var err2 error

	if temp == true {

		results, err2 = db.Exec(`UPDATE chomp_users SET password_hash=?, is_password_temp = ?, password_expiry = ?
							  WHERE chomp_user_id=?`, userInfo.PasswordHash, true, 
							  time.Now().Unix() + 86400, userInfo.UserID)
	} else {

		results, err2 = db.Exec(`UPDATE chomp_users SET password_hash=?, is_password_temp = ?, password_expiry = ?
							  WHERE chomp_user_id=?`, userInfo.PasswordHash, false, 0, userInfo.UserID)
	}
	
	id, err2 := results.LastInsertId()
	fmt.Printf("Results = %v\n err3 = %v\n", id , err2)
	return err2
}

func (userInfo UserInfo) UpdateInstaCode() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Printf("userinfo = %v\n", userInfo)
	fmt.Printf("Type of userInfo = %v\n", reflect.TypeOf(userInfo))
	var results sql.Result
	var err2 error

	results, err2 = db.Exec(`UPDATE chomp_users SET insta_code=?
							  WHERE chomp_username=?`, userInfo.InstaCode, userInfo.Username)

	if err2 != nil {
		fmt.Printf("Err = %v\n", err2)
		return err2
	}
	
	id, err2 := results.LastInsertId()
	fmt.Printf("Results = %v\n err3 = %v\n", id , err2)
	return err2
}

func (photo Photos) SetMePhoto() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", photo)
	fmt.Printf("Type of userInfo = %v\n", reflect.TypeOf(photo))

	query := fmt.Sprintf("INSERT into photos(chomp_user_id, file_path, file_hash, uuid) SELECT chomp_user_id, '%s', '%s', '%s' from chomp_users WHERE chomp_username='%s'", 
						photo.FilePath, photo.FileHash, photo.Uuid, photo.Username)
	fmt.Printf("Query = %v\n", query)

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
	row := db.QueryRow("SELECT id, chomp_user_id, file_path, file_hash, UNIX_TIMESTAMP(last_updated), uuid from photos where uuid=?", photo.Uuid).Scan(&photo.ID, &photo.UserID, &photo.FilePath, &photo.FileHash, &photo.TimeStamp, &photo.Uuid)
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

	err = db.QueryRow(`SELECT id, chomp_users.chomp_user_id, file_path, file_hash, UNIX_TIMESTAMP(last_updated), uuid
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

	err = db.QueryRow(`SELECT chomp_user_id, file_path, file_hash, UNIX_TIMESTAMP(last_updated), uuid
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

func (userInfo *UserInfo) DeleteAllPhotos() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", userInfo)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(userInfo))

	_, err = db.Query("DELETE FROM photos WHERE user_id=?", userInfo.UserID)

	return err
}

func (userInfo *UserInfo) AbandonAllPhotos() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", userInfo)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(userInfo))

	_, err = db.Query("UPDATE photos SET chomp_user_id = 0 WHERE chomp_user_id=?", userInfo.UserID)

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
	fmt.Printf("locaiton num = %v\n", restaurant.LocationNum)

	rows, err2 := db.Query(`SELECT id, name, latitude, longitude, location_num, source, source_location_id
						FROM restaurants
						WHERE name=?`,restaurant.Name)

	if err2 == nil  {

		for rows.Next() {
    		var id int
    		var name string
    		var latt float64
    		var long float64
    		var locationNum int
    		var source string
    		var sourceLocID string
    		err = rows.Scan(&id, &name, &latt, &long, &locationNum, 
    						&source, &sourceLocID)
    		fmt.Printf("locaiton num = %v\n", restaurant.LocationNum)
    		fmt.Printf("db Location num = %v\n", locationNum)
    		fmt.Printf("db restaurant id num = %v\n", id)
    		if locationNum >= restaurant.LocationNum {
    			restaurant.Latt = latt
    			restaurant.Long = long
    			restaurant.LocationNum = locationNum
    		}
    		restaurant.ID = id
    		restaurant.Name = name
    		restaurant.Source = source
    		restaurant.SourceLocID = sourceLocID
	
		}
	}
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
	fmt.Printf("Type of Restaurant = %v\n\n", reflect.TypeOf(restaurant))
	fmt.Printf(`INSERT INTO restaurants
						 SET id = %v, name = %v, latitude = %v, longitude = %v,
						 location_num = %v, source = %v,
						 source_location_id = %v`, restaurant.ID, restaurant.Name,
						 					      restaurant.Latt, restaurant.Long,
						 	  					  restaurant.LocationNum, restaurant.Source,
						 	  					  restaurant.SourceLocID)

	results, err2 := db.Exec(`INSERT INTO restaurants
						 SET name = ?, latitude = ?, longitude = ?,
						 location_num = ?, source = ?,
						 source_location_id = ?`, restaurant.Name,
						 					      restaurant.Latt, restaurant.Long,
						 	  					  restaurant.LocationNum, restaurant.Source,
						 	  					  restaurant.SourceLocID)
	id, err2 := results.LastInsertId()
	restaurant.ID = int(id)

	fmt.Printf("Results = %v\n err3 = %v\n", id , err2)

	return err2
}

func (restaurant *Restaurants) UpdateRestaurant() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("inside call: restaurants = %v\n", restaurant)
	fmt.Printf("Type of Restaurant = %v\n\n", reflect.TypeOf(restaurant))
	fmt.Printf(`UPDATE restaurants
				SET latitude = %v, longitude = %v,
				location_num = %v, source = %v,
				source_location_id = %v
				WHERE id = %v`, restaurant.Latt, restaurant.Long,
						 	  					  restaurant.Source,
						 	  					  restaurant.SourceLocID, restaurant.ID)

	results, err2 := db.Exec(`UPDATE restaurants
						 SET latitude = ?, longitude = ?,
						 location_num = ?, source = ?,
						 source_location_id = ?`, restaurant.Latt, restaurant.Long,
						 	  					  restaurant.LocationNum, restaurant.Source,
						 	  					  restaurant.SourceLocID)
	id, err2 := results.LastInsertId()
	restaurant.ID = int(id)

	fmt.Printf("Results = %v\n err3 = %v\n", id , err2)

	return err2
}


func GetReviewsByUserID(userId int) (reviews []Review) {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return reviews
	}
	defer db.Close()
	fmt.Printf("id = %v\n", userId)
	fmt.Printf(`SELECT reviews.id, reviews.user_id, reviews.username,
						dish_id, dish.name, photo_id, restaurant_id, restaurants.name,
						latitude, longitude, location_num, restaurants.source, source_location_id,
						price, liked, finished, description,
						UNIX_TIMESTAMP(reviews.created_date), UNIX_TIMESTAMP(reviews.last_updated), reviews.dish_tags,
						reviews.dish_tags2, reviews.dish_tag_ids
					   FROM reviews
					   JOIN restaurants on reviews.restaurant_id = restaurants.id
					   JOIN dish on reviews.dish_id = dish.id
					   WHERE user_id =%v\n` + "\n",userId)

		rows, err := db.Query(`SELECT reviews.id, reviews.user_id, reviews.username,
						dish_id, dish.name, photo_id, restaurant_id, reviews.source, restaurants.name,
						latitude, longitude, location_num, restaurants.source, source_location_id,
						price, liked, finished, description,
						UNIX_TIMESTAMP(reviews.created_date), UNIX_TIMESTAMP(reviews.last_updated), UNIX_TIMESTAMP(reviews.finished_time),
						reviews.dish_tag_ids
					   FROM reviews
					   JOIN restaurants on reviews.restaurant_id = restaurants.id
					   JOIN dish on reviews.dish_id = dish.id
					   WHERE user_id =?`,userId)

	if err != nil {
		fmt.Printf("Error while retrieving dish..%v\n", err)
		return reviews
	}
	var review Review
	var blobTags string
	var blobIds string

	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&review.ID, &review.UserID, &review.Username,
			&review.Dish.ID, &review.Dish.Name, &review.Photo.ID, &review.Restaurant.ID, &review.Source,
			&review.Restaurant.Name, &review.Restaurant.Latt, &review.Restaurant.Long, &review.Restaurant.LocationNum,
			&review.Restaurant.Source, &review.Restaurant.SourceLocID, &review.Price, &review.Liked, &review.Finished, &review.Description,
			&review.CreatedDate, &review.LastUpdated, &review.FinishedTime, &blobIds); err != nil {
			fmt.Printf("Err= %v\n", err.Error())
			return reviews
		}
		fmt.Printf("in for, review = %v\n", review)
		fmt.Printf("tags = %v\n", blobTags)

		blobIdSlice := strings.Fields(strings.Trim(blobIds, "[]"))

		var dishTag DishTag
		var newDishTagArray []DishTag

		for i, e := range blobIdSlice {
			fmt.Printf("dishTag IDs = %v: %v\n", i, e)
			id, err := strconv.Atoi(e)
			if err != nil {
				fmt.Printf("Error converting..%v\n", err)
				return reviews
			}
			rows, err := db.Query(`SELECT id, tag
					   FROM dish_tags
					   WHERE id =?`, id)
			if err != nil {
				fmt.Printf("Err= %v\n", err.Error())
				return reviews
			}
			defer rows.Close()
			for rows.Next() {
				if err := rows.Scan(&dishTag.ID, &dishTag.Tag); err != nil {
					fmt.Printf("Err= %v\n", err.Error())
					return reviews
				}
				review.DishTags = append(review.DishTags, dishTag)
				fmt.Printf("dishTag = %v\n", dishTag)
			}

		}
		fmt.Printf("ids = \n")
		fmt.Printf("newDishTagArray = %v\n", newDishTagArray)
		reviews = append(reviews, review)
	}
	fmt.Printf("\nReturning = %v\n", reviews)
	return reviews
}

func (review *Review) CreateReview() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Printf("REVIEW = %v\n", review)
	fmt.Printf("Type of review = %v\n", reflect.TypeOf(review))

	fmt.Printf("INSERT INTO reviews SET user_id = %v, username = %v, dish_id = %v, photo_id = %v, restaurant_id = %v, price = %v, liked = %v, dish_tags = %v, description = %v, finished = %v\n\n", 
												  review.UserID, review.Username,
						 					      review.Dish.ID, review.Photo.ID,
						 	  					  review.Restaurant.ID, review.Price,
						 	  					  review.Liked,review.DishTags, review.Description, review.Finished)
	fmt.Printf("Distags = %v\n", review.DishTags)
	fmt.Printf("Liked = %v\n", review.Liked)
	dishTagIds, err := review.AddDishTags()

	if err != nil {
		return err
	}

	var results sql.Result

	if review.Finished.Valid == true && review.Finished.Bool == true {
			results, err = db.Exec(`INSERT INTO reviews
						 SET user_id = ?, username = ?, dish_id = ?, dish_tag_ids=?,
						 photo_id = ?, restaurant_id = ?, price = ?,
						 liked = ?, finished = ?, description = ?, finished_time=UNIX_TIMESTAMP(now()),
						 source = ?`, review.UserID, review.Username,
						 					      review.Dish.ID,
						 					      fmt.Sprintf("%+v", dishTagIds),
						 	  					  review.Photo.ID, review.Restaurant.ID, 
						 	  					  review.Price, review.Liked, review.Finished,
						 	  					  review.Description, review.Source)
	} else {
		results, err = db.Exec(`INSERT INTO reviews
						 SET user_id = ?, username = ?, dish_id = ?, dish_tag_ids=?,
						 photo_id = ?, restaurant_id = ?, price = ?,
						 liked = ?, finished = ?, description = ?, source = ?`, review.UserID, review.Username,
						 					      review.Dish.ID,
						 					      fmt.Sprintf("%+v", dishTagIds),
						 	  					  review.Photo.ID, review.Restaurant.ID, 
						 	  					  review.Price, review.Liked, review.Finished,
						 	  					  review.Description, review.Source)
	}

	if err != nil {
		fmt.Printf("Error = %v", err)
		return err
	}
	id, err := results.LastInsertId()
	review.ID = int(id)	

	fmt.Printf("Results = %v\n err3 = %v\n", id , err)
	return err
}

func (review *Review) GetReviewLastTimeStamp(reviewId int) error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Printf("REVIEW = %v\n", review)
	fmt.Printf("Type of review = %v\n", reflect.TypeOf(review))

	fmt.Printf("SELECT last_updated from reviews WHERE id = %v\n", reviewId)
	fmt.Printf("Distags = %v\n", review.DishTags)
	fmt.Printf("Liked = %v\n", review.Liked)
	if err != nil {
		return err
	}
	
	err = db.QueryRow(`SELECT UNIX_TIMESTAMP(last_updated) from reviews WHERE id = ?`, reviewId).Scan(&review.LastUpdated)

	if err != nil {
		fmt.Printf("Error = %v", err)
		return err
	}
	return err
}


func (review *Review) AddDishTags() ([]int, error) {
// func (review *Review) AddDishTags() ([]DishTag, error) {

	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return make([]int, 0), err
		// return make([]DishTag, 0), err
	}
	defer db.Close()

	var dishTagIds []int
	// var dishTags []DishTag
	// for _, e := range review.DishTags {
	for _, e := range review.DishTags {

		fmt.Printf("Insert Dishtags %v\n", e)
		results, err := db.Exec(`INSERT INTO dish_tags
						 			SET tag = ?, count = count+1
						 			ON DUPLICATE KEY UPDATE count = count+1`, e.Tag)

		if err != nil {
			fmt.Printf("Error = %v", err)
			return dishTagIds, err
			// return dishTags, err
		}

		id, err := results.LastInsertId()
		// id, err := results.LastInsertId()
		if err != nil {
			fmt.Printf("Error = %v", err)
			return dishTagIds, err
			// return dishTags, err
		}
		dishTagIds = append(dishTagIds, int(id))
		// e.ID = int(id)
		// dishTags = append(dishTags, e)
	}
	return dishTagIds, nil
	// return dishTags, nil
}

func (igStore *IgStore) UpdateLastPull() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Printf("UserId = %v\n", igStore.UserID)
	fmt.Printf("media ID %v\n", igStore.IgMediaID)
	fmt.Printf("All = %v\n", igStore)

	fmt.Printf(`\nINSERT INTO ig_last_crawl
				SET user_id = %v, ig_media_id = %v, epoch_now = %v, ig_created_timestamp = %v
				ON DUPLICATE KEY UPDATE ig_media_id = %v, epoch_now = %v, ig_created_timestamp = %v\n`,
				igStore.UserID, igStore.IgMediaID, time.Now().Unix(), 
				igStore.IgCreatedTime, igStore.IgMediaID, time.Now().Unix(), 
				igStore.IgCreatedTime)

	results, err2 := db.Exec(`INSERT INTO ig_last_crawl
						 SET user_id = ?, ig_media_id = ?, epoch_now = ?, ig_created_timestamp = ?
						 ON DUPLICATE KEY UPDATE ig_media_id = ?, epoch_now = ?, ig_created_timestamp = ?`,
						 igStore.UserID, igStore.IgMediaID, time.Now().Unix(), 
						 igStore.IgCreatedTime, igStore.IgMediaID, time.Now().Unix(), 
						 igStore.IgCreatedTime)

	if err2 != nil {
		fmt.Printf("Error = %v", err2)
		return err2
	}
	id, err2 := results.LastInsertId()

	fmt.Printf("Results = %v\n err3 = %v\n", id , err2)
	return err2
}

func (igStore *IgStore) GetLastPull() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Printf("UserId = %v\n", igStore.UserID)
	fmt.Printf("media ID %v\n", igStore.IgMediaID)


	err2 := db.QueryRow(`SELECT user_id, ig_media_id, epoch_now, ig_created_timestamp
						 FROM ig_last_crawl
						 WHERE user_id = ?`, igStore.UserID).Scan(&igStore.UserID,
						 									&igStore.IgMediaID, 
						 									&igStore.Epoch, 
						 									&igStore.IgCreatedTime)

	if err2 != nil {
		fmt.Printf("Error = %v\n", err2)
		return err2
	}

	fmt.Println("Inside DB: pull now: ", igStore)
	return err2
}



func (review *Review) UpdateReview() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Printf("In Update Rewviews, Review = %v\nreview Id = %v\n", 
		reflect.TypeOf(review), review.ID)

	fmt.Printf("Distags = %v\n", review.DishTags)
	fmt.Printf("Liked = %v\n", review.Liked)
	dishTags, err := review.AddDishTags()

	if err != nil {
		return err
	}

	dishTagsCr := fmt.Sprintf("%+v",dishTags)
	fmt.Printf("DishTagCr = %v\n", dishTagsCr)

	fmt.Printf(`UPDATE reviews
						 SET user_id = %v, username = %v, dish_id = %v, dish_tags2 = %v,
						 photo_id = %v, restaurant_id = %v, price = %v,
						 liked = %v, finished = %v, description = %v
						 WHERE id = %v\n\n`, review.UserID, review.Username,
						 					      review.Dish.ID, dishTagsCr, //dishTagIdsCr,
						 	  					  review.Photo.ID, review.Restaurant.ID, 
						 	  					  review.Price, review.Liked, review.Finished,
						 	  					  review.Description, review.ID)

	dishTagIds, err := review.AddDishTags()
	if err != nil {
		return err
	}
	var results sql.Result

	if review.Finished.Valid  == true && review.Finished.Bool == true  {
		results, err = db.Exec(`UPDATE reviews
					 SET user_id = ?, username = ?, dish_id = ?, dish_tag_ids=?,
					 photo_id = ?, restaurant_id = ?, price = ?,
					 liked = ?, finished = ?, description = ?, finished_time=UNIX_TIMESTAMP(now()),
					 source = ?
					 WHERE id = ?`, review.UserID, review.Username,
					 					      review.Dish.ID,
					 					      fmt.Sprintf("%+v", dishTagIds),
					 	  					  review.Photo.ID, review.Restaurant.ID, 
					 	  					  review.Price, review.Liked, review.Finished,
					 	  					  review.Description, review.Source, review.ID)
		
	} else {
		results, err = db.Exec(`UPDATE reviews
					 SET user_id = ?, username = ?, dish_id = ?, dish_tag_ids=?,
					 photo_id = ?, restaurant_id = ?, price = ?,
					 liked = ?, finished = ?, description = ?, source = ?
					 WHERE id = ?`, review.UserID, review.Username,
					 					      review.Dish.ID,
					 					      fmt.Sprintf("%+v", dishTagIds),
					 	  					  review.Photo.ID, review.Restaurant.ID, 
					 	  					  review.Price, review.Liked, review.Finished,
					 	  					  review.Description, review.Source, review.ID)
	}


	if err != nil {
		fmt.Printf("Error = %v", err)
		return err
	}
	rows, err := results.RowsAffected()
	if rows < 1 {
		fmt.Printf("Nothing updated, dec counter\n")
		err = errors.New("0 rows updated, Might be duplicate")
		DecDishTagCounter(dishTagIds)
	}

	return err
}

func DecDishTagCounter(dishTagIds []int) (error) {

	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	fmt.Printf("DecDishCounter %v\n", dishTagIds)

	for _, e := range dishTagIds {

		fmt.Printf("Decrement Dishtags %v\n", e)
		_, err := db.Exec(`UPDATE dish_tags
						 	SET count = count-1
						 	WHERE id = ?`, e)

		if err != nil {
			fmt.Printf("Error = %v", err)
			return err
		}
	}
	return nil
}


func (review *Review) DeleteReview() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", review)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(review))

	results, err2 := db.Exec("DELETE FROM reviews WHERE id=?", review.ID)
	if err2 != nil {
		fmt.Printf("Error = %v", err2)
		return err2
	}
	rows, err2 := results.RowsAffected()
	if rows < 1 {
		fmt.Printf("Nothing updated\n")
		err2 = errors.New("0 rows deleted")
	}

	return err2
}

func (userInfo *UserInfo) DeleteAllReviewsByUser() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", userInfo)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(userInfo))

	results, err2 := db.Exec("DELETE FROM reviews WHERE user_id=?", userInfo.UserID)
	if err2 != nil {
		fmt.Printf("Error = %v", err2)
		return err2
	}
	rows, err2 := results.RowsAffected()
	if rows < 1 {
		fmt.Printf("Nothing deleted\n")
		err2 = errors.New("0 rows deleted")
	}

	return err2
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

func Logout(sessionId string) error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Printf("SessionId = %v\n", sessionId)
	fmt.Printf("Time Now = %v\n", time.Now().Unix())

	_, err = db.Query("UPDATE session SET session_expiry=? WHERE session_key=?", time.Now().Unix() - globalsessionkeeper.ChompConfig.ManagerConfig.Maxlifetime, sessionId)

	if err != nil {
		fmt.Printf("Got an error. %v\n", err.Error())
	}

	return err
}

func LogoutAllSessions(username string) error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Print("Logging out user = %v\n", username)

	rows, err := db.Query("SELECT * FROM session WHERE session_data LIKE ?", "%" + username + "%")

	if err == nil  {
		var sessionData []byte
		var sessionKey string
		var sessionExpiry int64

		for rows.Next() {

			rows.Scan(&sessionKey, &sessionData, &sessionExpiry)
			kv, err := session.DecodeGob(sessionData)

			if err != nil {

				fmt.Printf("Error scaning..%v\n", err.Error())
				return err
			}

			fmt.Printf("\n\nkv = %v\n\n", kv)
			if kv["username"] == username {
				Logout(sessionKey)
			}
		}
	}

	return err
}
