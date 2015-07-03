package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"errors"
	"time"
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
	TimeStamp	string			`json:"timeStamp"`
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
// type DishTags struct {
// 	Tags 	[]string
// }
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
	DishTags		string 	 		`json:"dishTags"`
	CreatedDate		string 			`json:"createdDate,omitempty"`
	LastUpdated 	string 			`json:"lastUpdated,omitempty"`
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
	LastUpdated 	string
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
						is_password_temp, password_expiry, fname, lname
					   FROM chomp_users
					   WHERE chomp_username=?`, 
					   userInfo.Username).Scan(&userInfo.UserID, &userInfo.Email,
					   							    &userInfo.Username, &userInfo.PhoneNumber,
					   							    &userInfo.PasswordHash,&userInfo.DOB,
					   							    &userInfo.Gender, &userInfo.Photo.ID,
					   							    &userInfo.IsPasswordTemp, &userInfo.PasswordExpiry,
					   							    &userInfo.Fname, &userInfo.Lname)
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
	fmt.Printf("SELECT chomp_user_id, email, chomp_username,phone_number, password_hash, dob, gender, photo_id,fname = ?, lname = ? FROM chomp_users WHERE chomp_username=%s\n", userInfo.Username)

	err = db.QueryRow(`SELECT chomp_user_id, email, chomp_username,
						dob, gender, photo_id, is_password_temp, password_expiry,
						fname, lname
					   FROM chomp_users
					   WHERE email=?`, 
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

func (photo Photos) SetMePhoto() error {
	db, err := sql.Open("mysql", "root@tcp(172.16.0.1:3306)/chomp")
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare statement for writing chomp_users table data
	fmt.Println("map = %v\n", photo)
	fmt.Printf("Type of userInfo = %v\n", reflect.TypeOf(photo))

	//query := fmt.Sprintf("INSERT INTO photos SET dish_id='%d', chomp_user_id='%d', file_path='%s', file_hash='%s', uuid='%s'", photo.DishID, photo.UserID, photo.FilePath, photo.FileHash, photo.Uuid)
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
	row := db.QueryRow("SELECT id, chomp_user_id, file_path, file_hash, last_updated, uuid from photos where uuid=?", photo.Uuid).Scan(&photo.ID, &photo.UserID, &photo.FilePath, &photo.FileHash, &photo.TimeStamp, &photo.Uuid)
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

	err = db.QueryRow(`SELECT id, chomp_users.chomp_user_id, file_path, file_hash, last_updated, uuid
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

	err = db.QueryRow(`SELECT chomp_user_id, file_path, file_hash, last_updated, uuid
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
	fmt.Printf("id = %v", userId)
	rows, err := db.Query(`SELECT reviews.id, reviews.user_id, reviews.username,
						dish_id, dish.name, photo_id, restaurant_id, restaurants.name,
						latitude, longitude, location_num, source, source_location_id,
						price, liked, finished, description,
						reviews.created_date, reviews.last_updated
					   FROM reviews
					   JOIN restaurants on reviews.restaurant_id = restaurants.id
					   JOIN dish on reviews.dish_id = dish.id
					   WHERE user_id =?`,userId)
	if err != nil {
		return reviews
	}
	var review Review
	// reviews := []Review{}
	for rows.Next() {
		if err := rows.Scan(&review.ID, &review.UserID, &review.Username,
			&review.Dish.ID, &review.Dish.Name, &review.Photo.ID, &review.Restaurant.ID,
			&review.Restaurant.Name, &review.Restaurant.Latt, &review.Restaurant.Long, &review.Restaurant.LocationNum,
			&review.Restaurant.Source, &review.Restaurant.SourceLocID, &review.Price, &review.Liked, &review.Finished, &review.Description,
			&review.CreatedDate, &review.LastUpdated); err != nil {
			fmt.Printf("Err= %v\n", err.Error())
			return reviews
		}
		fmt.Printf("in for, review = %v\n", review)
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

	fmt.Printf("INSERT INTO reviews SET user_id = %v, username = %v, dish_id = %v, photo_id = %v, restaurant_id = %v, price = %v, liked = %v, dish_tags = %v, description = %v\n\n", 
												  review.UserID, review.Username,
						 					      review.Dish.ID, review.Photo.ID,
						 	  					  review.Restaurant.ID, review.Price,
						 	  					  review.Liked,review.DishTags, review.Description)
	fmt.Printf("Distags = %v\n", review.DishTags)

	results, err2 := db.Exec(`INSERT INTO reviews
						 SET user_id = ?, username = ?, dish_id = ?, dish_tags=?,
						 photo_id = ?, restaurant_id = ?, price = ?,
						 liked = ?, finished = ?, description = ?`, review.UserID, review.Username,
						 					      review.Dish.ID, review.DishTags,
						 	  					  review.Photo.ID, review.Restaurant.ID, 
						 	  					  review.Price, review.Liked, review.Finished,
						 	  					  review.Description)

	if err2 != nil {
		fmt.Printf("Error = %v", err2)
		return err2
	}
	id, err2 := results.LastInsertId()
	review.ID = int(id)	

	fmt.Printf("Results = %v\n err3 = %v\n", id , err2)
	return err2
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
	fmt.Println("map = %v\n", review)
	fmt.Print("Type of userInfo = %v\n", reflect.TypeOf(review))

	results, err2 := db.Exec(`UPDATE reviews
						 SET user_id = ?, username = ?, dish_id = ?, dish_tags=?,
						 photo_id = ?, restaurant_id = ?, price = ?,
						 liked = ?, finished = ?, description = ? WHERE id = ?`, review.UserID, review.Username,
						 					      review.Dish.ID, review.DishTags, review.Photo.ID,
						 	  					  review.Restaurant.ID, review.Price, review.Liked,
						 	  					  review.Finished, review.Description, review.ID)
	if err2 != nil {
		fmt.Printf("Error = %v", err2)
		return err2
	}
	rows, err2 := results.RowsAffected()
	if rows < 1 {
		fmt.Printf("Nothing updated\n")
		err2 = errors.New("0 rows updated")
	}

	return err2
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

func IsValid(s sql.NullString) string {

	if s.Valid {
		fmt.Println("s is valid")
		return s.String
	} else {
		fmt.Println("s is not valid")
		return s.String
	}
}
