package review

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/parnurzeal/gorequest"
	"chompapi/db"
	"chompapi/globalsessionkeeper"
	"reflect"
	"strings"
	"chompapi/me"
	"strconv"
	"database/sql"
)

type ParentData struct {
	Data 		[]InstaData
}

type InstaData struct {
	Tags			[]string
	Type 			string
	Location 		Location
	Comments 		Comments
	filter 			string
	Created_Time 	string
	Link			string
	Likes 			Likes
	Images 			Images
	Caption 		Caption
	User_Has_Liked	bool
	ID 				string
	User 			User
}

type Location struct {
	ID 				int64
	Latitude		float64
	Name 			string
	Longitude 		float64
}

type Images struct {
	Low_Resolution 		Res
	Thumbnail			Res
	Standard_Resolution 	Res
}

type Res struct {
	Url 		string
	Width 		int
	Height 		int
}

type Caption struct {
	ID 				string
	Created_Time 	string
	Text 			string
	From			User
}

type Likes struct {
	Count			int
	Data 			[]User
}

type Comments struct {
	Count 			int
	Data 			[]Data
}

type Data struct {
	ID 				string
	Text 			string
	From 			User
}

type User struct {
	ID 				string
	Username 		string
	ProfilePicture	string
	FullName 		string
}

func Crawl(w http.ResponseWriter, r *http.Request) {
	var myErrorResponse globalsessionkeeper.ErrorResponse
	cookie := globalsessionkeeper.GetCookie(r)
	if cookie == "" {
			//need logging here instead of print
		fmt.Println("Cookie = %v", cookie)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	sessionStore, err := globalsessionkeeper.GlobalSessions.GetSessionStore(cookie)
	if err != nil {
			//need logging here instead of print
			w.WriteHeader(http.StatusUnauthorized)
			return
	}
	//input.Username = sessionStore.Get("username")
	sessionUser := sessionStore.Get("username")
	sessionUserID := sessionStore.Get("userId")
	fmt.Println("SessionUser = %v", sessionUser)

	if sessionUser == nil {
			//need logging here instead of print
			fmt.Printf("Username not found, returning unauth, Get has %v\n", sessionStore)
			w.WriteHeader(http.StatusUnauthorized)
			return
	} else {
		//reset time to time.now() + maxlifetime
		defer sessionStore.SessionRelease(w)

		//create variables
		username 	 := reflect.ValueOf(sessionUser).String()
		userId 	 	 := reflect.ValueOf(sessionUserID).Int()
		crawl 	 	 := new(db.Crawl)
		instaData 	 := new(ParentData)
		// token 		 := "1695698585.a60a4c1.e82730bfc557441e937d84ccc2aa1d99"
		// instaAuthUrl := "https://instagram.com/oauth/authorize/?client_id=a60a4c1bc76f45108e75a9d09b566832&redirect_uri=http://www.thechompapp.com/&response_type=token"
		instaRMediaUrl := "https://api.instagram.com/v1/users/self/media/recent/?access_token=%v"

		crawl.Username = username
		crawl.UserID = int(userId)

		switch r.Method {

		case "POST":
			decoder := json.NewDecoder(r.Body)
			// fmt.Printf("r.Body = %v\n", reflect.ValueOf(r.Body.Value()).String())
			if err := decoder.Decode(&crawl); err != nil {
				//need logging here instead of print
				fmt.Printf("something went wrong in login %v", err)
				myErrorResponse.Code = http.StatusBadRequest
				myErrorResponse.Error = "Malformed JSON: " + err.Error()
				myErrorResponse.HttpErrorResponder(w)
				return
			}
			url :=  fmt.Sprintf(instaRMediaUrl, crawl.InstaTok)
			request := gorequest.New()
			resp, body, errs := request.Get(url).End()
			if errs != nil {
				fmt.Printf("something went wrong in get %v", err)
				myErrorResponse.Code = http.StatusBadRequest
				myErrorResponse.Error = "Malformed JSON: " + err.Error()
				myErrorResponse.HttpErrorResponder(w)
				return
			}
			fmt.Printf("Type of body = %v\n", reflect.TypeOf(body))

			// var objmap map[string]*json.RawMessage
			//instaData,err := json.Marshal(body)

			//emptyString := make([]string, 0)
			//var newW io.Writer
            //err := json.NewEncoder(newW).Encode(&body)
            err := json.Unmarshal([]byte(body), &instaData)

			if err != nil {
				fmt.Printf("Err = %v", err)
			}
			// fmt.Printf("Obj = %v\n", objmap["data"])
			fmt.Printf("Resp:%v \nbody: %v\n, errs: %v\n", resp, body, errs)
			fmt.Printf("instaData = %v\n", instaData.Data[0])
			fmt.Printf("instaData images = %v\n", instaData.Data[0].Images)
			fmt.Printf("instaData comments = %v\n", instaData.Data[0].Comments)
			fmt.Printf("instaData tags = %v\n", instaData.Data[0].Tags)
			//review = new(db.Review)
			for index, each := range instaData.Data {
				fmt.Printf("Index = %v\neach = %v\n", index, each)
				for i, e := range each.Tags {
					fmt.Printf("Tag %v: %s\n", i,e)
					if strings.Contains(strings.ToLower(e), "chomp") {
						fmt.Println("Contains chomp")
						//generate and store UUID for photo
						photoInfo := CreatePhoto(username)
						if photoInfo.ID == 0 {
							fmt.Println("Something went wrong to create photo")
						}
						//upload file to google storeage

						//update database
						err := each.CreateReview(photoInfo)
						if err == nil {
							break
						} else {
							fmt.Println("No Review Created")
						}
					}
				}
			}


		default:
			myErrorResponse.Code = http.StatusMethodNotAllowed
			myErrorResponse.Error = err.Error()
			myErrorResponse.HttpErrorResponder(w)
		}
	}
}

func CreatePhoto(username string) db.Photos {
	var photoInfo db.Photos

	photoInfo.Uuid = me.GenerateUuid()
	photoInfo.Username = username
	
	err := photoInfo.SetMePhoto()
	if err != nil {
		//need logging here instead of print
		// myErrorResponse.Code = http.StatusInternalServerError
		// myErrorResponse.Error = err.Error()
		// myErrorResponse.HttpErrorResponder(w)
		return photoInfo
	} 
	err2 := photoInfo.GetPhotoInfoByUuid()
	if err2 != nil {
		//need logging here instead of print
		fmt.Printf("Something went wrong in db, %v\n", err2)
		// myErrorResponse.Code = http.StatusInternalServerError
		// myErrorResponse.Error = err.Error()
		// myErrorResponse.HttpErrorResponder(w)
		return photoInfo
	}
	return photoInfo
}

func (instaData *InstaData) CreateReview(photoInfo db.Photos) error {
	review := new(db.Review)
	dbRestaurant := new(db.Restaurants)
	//fill in restaurant info
	review.UserID = photoInfo.UserID
	review.Username = photoInfo.Username
	review.Restaurant.Name = instaData.Location.Name
	review.Restaurant.Latt = instaData.Location.Latitude
	review.Restaurant.Long = instaData.Location.Longitude
	review.Restaurant.Source = "instagram"
	review.Restaurant.SourceLocID = strconv.FormatInt(instaData.Location.ID, 10)

	dbRestaurant.Name = instaData.Location.Name
	err := dbRestaurant.GetRestaurantInfoByName()
	if err != nil && err != sql.ErrNoRows{
		//something bad happened
		fmt.Printf("something went while retrieving data %v", err)
		// myErrorResponse.Code = http.StatusInternalServerError
		// myErrorResponse.Error = "something went while retrieving data:-:" + err2.Error()
		// myErrorResponse.HttpErrorResponder(w)
		return err
	} else if err == sql.ErrNoRows || dbRestaurant.ID == 0 {
		// not found in DB
		if review.Restaurant.Name != "" {
			fmt.Println("Restaurant Not found in DB, creating new entry")

			err = review.Restaurant.CreateRestaurant()
			if err != nil {
				//something bad happened
				fmt.Printf("something went while retrieving data %v", err)
				// myErrorResponse.Code = http.StatusInternalServerError
				// myErrorResponse.Error = "something went while retrieving data:-:" + err.Error()
				// myErrorResponse.HttpErrorResponder(w)	
				return err
			}
		} else {
			// Restaurant Value Blank
			fmt.Println("Blank Restaurant found in db")
			fmt.Println("Blank Restaurant In DB", dbRestaurant)
			review.Restaurant = *dbRestaurant
		
		}
	} else {
		// entry found in db
		fmt.Println("Restaurant found in db")
		fmt.Println("Restaurant In DB", dbRestaurant)
		if review.Restaurant.Source == dbRestaurant.Source {
			//same source, check location ID for same restaurnt
			fmt.Println("same source")
			if review.Restaurant.SourceLocID != dbRestaurant.SourceLocID {
				//creaet new restaurant with +1 to location_num
				fmt.Println("location id !=")
				review.Restaurant.LocationNum = dbRestaurant.LocationNum + 1
				err = review.Restaurant.CreateRestaurant()
				if err != nil {
					//something bad happened
					fmt.Printf("something went while retrieving data %v", err)
					// myErrorResponse.Code = http.StatusInternalServerError
					// myErrorResponse.Error = "something went while retrieving data:-:" + err.Error()
					// myErrorResponse.HttpErrorResponder(w)
					return err
				}
			} else {
				//use existing DB values
				fmt.Println("Location ID Equal, using db values")
				review.Restaurant = *dbRestaurant
			}
		} else if dbRestaurant.Source == "instagram"  {
			//trust DB over New
			fmt.Println("Source not same, DB == insta")
			review.Restaurant = *dbRestaurant
			//review.CreateReview()
		} else if review.Restaurant.Source == "instagram" {
			fmt.Println("New restaurant instagram, updating db")
			if dbRestaurant.LocationNum == 0 {
				review.Restaurant.UpdateRestaurant()
				if err != nil {
					//something bad happened
					fmt.Printf("something went while retrieving data %v", err)
					// myErrorResponse.Code = http.StatusInternalServerError
					// myErrorResponse.Error = "something went while retrieving data:-:" + err.Error()
					// myErrorResponse.HttpErrorResponder(w)
					return err
				}
			} else {
				fmt.Println("location id !=")
				review.Restaurant.LocationNum = dbRestaurant.LocationNum + 1
				err = review.Restaurant.CreateRestaurant()
				if err != nil {
					//something bad happened
					fmt.Printf("something went while retrieving data %v", err)
					// myErrorResponse.Code = http.StatusInternalServerError
					// myErrorResponse.Error = "something went while retrieving data:-:" + err.Error()
					// myErrorResponse.HttpErrorResponder(w)
					return err
				}
			}
		} 
	}
	var tags string
	for _,e := range instaData.Tags {
		fmt.Printf("adding tag %v\n", e)
		if tags != "" {
			tags = tags + "," + e
		} else {
			tags = e
		}
	}
	review.DishTags = tags
	if instaData.Likes.Count > 0 {
		review.Liked.Bool = true
		// review.Liked.Value = true
	}
	review.Finished.Bool = false
	review.Description = instaData.Caption.Text

	// create review
	err = review.CreateReview()
	if err != nil {
		//something bad happened
		fmt.Printf("something went while retrieving data %v", err)
		// myErrorResponse.Code = http.StatusInternalServerError
		// myErrorResponse.Error = "could not create review:-:" + err.Error()
		// myErrorResponse.HttpErrorResponder(w)
		return err
	}
	return nil
}
