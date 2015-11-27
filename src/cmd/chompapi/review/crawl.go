package review

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/parnurzeal/gorequest"
	"cmd/chompapi/db"
	"cmd/chompapi/globalsessionkeeper"
	"reflect"
	"strings"
	"cmd/chompapi/me"
	"strconv"
	"database/sql"
	"cmd/chompapi/crypto"
	"golang.org/x/net/context"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/jwt"
    "google.golang.org/api/storage/v1"
    "io/ioutil"
    "os"
    "errors"
    "net/url"
    "io"
    "time"
    "golang.org/x/oauth2/jws"
    "regexp"
)

type ParentData struct {
	Data 		[]InstaData `json:"data"`
}

type InstaData struct {
	Tags			[]string 	`json:"tags"`
	Type 			string 		`json:"type"`
	Location 		Location 	`json:"location"`
	Comments 		Comments 	`json:"comments"`
	Filter 			string 		`json:"filter"`
	CreatedTime 	string 		`json:"created_time"`
	Link			string 		`json:"link"`
	Likes 			Likes 		`json:"likes"`
	Images 			Images 		`json:"images"`
	Caption 		Caption 	`json:"caption"`
	UserHasLiked	bool 		`json:"user_has_liked"`
	ID 				string 		`json:"id"`
	User 			User 		`json:"user"`
}

type Location struct {
	ID 				int64 		`json:"id"`
	Latitude		float64 	`json:"latitude"`
	Name 			string 		`json:"name"`
	Longitude 		float64 	`json:"longitude"`
}

type Images struct {
	LowRes		 	Res `json:"low_resolution"`
	Thumbnail		Res `json:"thumbnail"`
	StandardRes 	Res `json:"standard_resolution"`
}

type Res struct {
	Url 		string 	`json:"url"`
	Width 		int 	`json:"width"`
	Height 		int 	`json:"height"`
}

type Caption struct {
	ID 				string 	`json:"id"`
	Created_Time 	string 	`json:"created_time"`
	Text 			string 	`json:"text"`
	From			User 	`json:"from"`
}

type Likes struct {
	Count			int 	`json:"count"`
	Data 			[]User 	`json:"data"`
}

type Comments struct {
	Count 			int 	`json:"count"`
	Data 			[]Data 	`json:"data"`
}

type Data struct {
	ID 				string 	`json:"id"`
	Text 			string 	`json:"text"`
	From 			User 	`json:"from"`
}

type User struct {
	ID 				string 	`json:"id"`
	Username 		string 	`json:"username"`
	ProfilePicture	string 	`json:"profile_picture"`
	FullName 		string 	`json:"full_name"`
}

type GoogToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	IDToken     string `json:"id_token"`
	ExpiresIn   int64  `json:"expires_in"` // relative seconds from now
}

type StorageReq struct {
	//GoogToken 		GoogToken
	Token 			*oauth2.Token
	Bucket 			string
	FileName 		string
	FileUuid 		string
	FileLoc 		string
}

type ConfigFile struct {
    FileLoc		    string `json:"file_loc"`
    Bucket 			string `json:"bucket"`
}

var FileDownload ConfigFile

func Crawl(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {


	sessionUser := a.SessionStore.Get("username")
	sessionUserID := a.SessionStore.Get("userId")
	fmt.Println("SessionUser = %v", sessionUser)
	//create variables
	username 	 := reflect.ValueOf(sessionUser).String()
	userId 	 	 := reflect.ValueOf(sessionUserID).Int()

	crawl 	 	 := new(db.Crawl)
	instaData 	 := new(ParentData)
	igStore 	 := new(db.IgStore)

	// instaRMediaUrl 	:= "https://api.instagram.com/v1/users/self/media/recent/?access_token=%v&min_timestamp=%v"
	instaRMediaUrl 	:= "https://api.instagram.com/v1/users/self/media/recent/?access_token=%v&min_id=%v"

	crawl.Username = username
	crawl.UserID = int(userId)
	igStore.UserID = int(userId)

	switch r.Method {

	case "POST":

		// return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Temporarily Down"}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&crawl); err != nil {
			//need logging here instead of print
			fmt.Printf("something went wrong in login %v", err)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Malformed JSON: " + err.Error()}
		}
		fileContent, err := ioutil.ReadFile("./chomp_private/file_download.json")
		if err != nil {
    	    return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
    	}
    	err = json.Unmarshal(fileContent, &FileDownload)
    	if err != nil {
    	    fmt.Printf("Err = %v", err)
    	    return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Could not decode"}
    	}

    	/* //////////////////////////////////////// */
		/*                Check Last Crawl 			*/
		/* //////////////////////////////////////// */

		fmt.Println("=======================================")
		igStore.GetLastPull(a.DB)
		if err != nil {
			fmt.Printf("Error = %v\n")
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Could not set last crawl: " + err.Error()}
		}
		fmt.Printf("\nigStore Pull = %v\n", igStore)
		// fmt.Println("=======================================")

		var igMediaIdInt int64
		var igMediaId []string
		firstCrawl := true
		if igStore.IgMediaID != "fake" {
			firstCrawl := false
			igMediaId = copy(igMediaId, strings.Split(igStore.IgMediaID, "_"))
			igMediaIdInt, err = strconv.ParseInt(igMediaId[0], 10, 64)
	
			if err != nil {
				fmt.Printf("something went wrong while parsing ig media id %v", err)
				return globalsessionkeeper.ErrorResponse{http.StatusServiceUnavailable, err.Error()}
			}
		} else {
			igMediaIdInt = 0
		}

		iurl :=  fmt.Sprintf(instaRMediaUrl, crawl.InstaTok, strings.Join([]string{strconv.Itoa(int(igMediaIdInt + 1)), igMediaId[1]}, "_"))
		fmt.Printf("Media full = %v\n", igStore.IgMediaID)
		fmt.Printf("Media id p1 = %v\n", igMediaId[0])
		fmt.Printf("Media id p2 = %v\n", igMediaId[1])
		fmt.Printf("Media url = %v\n", iurl)
		fmt.Println("=======================================")
		request := gorequest.New()
		resp, body, errs := request.Get(iurl).End()

		if errs != nil {
			fmt.Printf("something went wrong in get %v", err)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Malformed JSON: " + err.Error()}
		}

        err = json.Unmarshal([]byte(body), &instaData)

		if err != nil {
			fmt.Printf("Err = %v", err)
			return globalsessionkeeper.ErrorResponse{http.StatusServiceUnavailable, "Communication Issues:IG: " + err.Error()}
		}

		if len(instaData.Data) == 0 {
			fmt.Println("No New Photos")
			return globalsessionkeeper.ErrorResponse{http.StatusNoContent, "Nothing to update"}
		}

		fmt.Printf("Resp:%v \nbody: %v\n, errs: %v\n", resp, body, errs)
		fmt.Printf("instaData Full = %v\n", instaData)
		fmt.Printf("instaData = %v\n", instaData.Data[0])
		fmt.Printf("instaData images = %v\n", instaData.Data[0].Images)
		fmt.Printf("instaData comments = %v\n", instaData.Data[0].Comments)
		fmt.Printf("instaData tags = %v\n", instaData.Data[0].Tags)

		var instaDataList []InstaData
		for _, tag_word := range crawl.Tags {
			fmt.Printf("-----Tag----- = %v\n", tag_word)
			for index, each := range instaData.Data {

				fmt.Printf("Index = %v\neach = %v\n", index, each)
				for i, e := range each.Tags {

					fmt.Printf("Tag %v: %s\n", i,e)
					if strings.Contains(strings.ToLower(e), tag_word) {

						fmt.Printf("\n\nContains %v\n", tag_word)
						fmt.Printf("Adding %v\n\n", index)
						// add to unique slice
						instaDataList = append(instaDataList, each)
					}
				}
			}
		}
		/*******************************************************************/
		/*                   SEND CRAWL TO DoCrawl()                       */
		/*******************************************************************/
		var reviews []review
		desc := "First Crawl"
		code := http.StatusNoContent
		if len(instaDataList) == 0 {
			fmt.Println("No New Photos")
			return globalsessionkeeper.ErrorResponse{http.StatusNoContent, "Nothing to update"}
		}
		if firstCrawl == false {
			desc, code, reviews, err = DoCrawl(a, username, &ParentData{instaDataList}, true)
			if err != nil {
				fmt.Printf("something went wrong in do crawl %v", err)
				return globalsessionkeeper.ErrorResponse{code, desc}
			}
		}
		/* //////////////////////////////////////// */
		/*               Set Last Crawl 			*/
		/* //////////////////////////////////////// */

		if len(instaDataList) > 0 {

			igStore.IgMediaID = instaDataList[0].ID
			// layOut := "Jan 2, 2006 at 3:04pm (MST)"
			// timeStamp, err := time.Parse(layOut, instaDataList[0].CreatedTime)

			// if err != nil {
			// 	fmt.Println(err)
			// 	return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "Not all reviews added: " + err.Error()}
			// }
			timeEpoch, err := strconv.ParseInt(instaDataList[0].CreatedTime, 10, 64)
			if err != nil {
				fmt.Println(err)
				return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "Last Crawl not Saved: " + err.Error()}
			}
			// igStore.IgCreatedTime = int(timeStamp.Unix())
			igStore.IgCreatedTime = int(timeEpoch)

			err = igStore.UpdateLastPull(a.DB)
	
			if err != nil {
				fmt.Printf("Could not update table\n")
				return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "Not all reviews added: " + err.Error()}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(reviews)
        if err != nil {
        	return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
        }
        fmt.Printf("Reviews = %v\n", reviews)
        w.WriteHeader(http.StatusOK)
		return nil

	default:

		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Method Not Allowed"}
	}
}

func CreatePhoto(username string, a *globalsessionkeeper.AppContext) db.Photos {

	var photoInfo db.Photos

	photoInfo.Uuid = me.GenerateUuid()
	photoInfo.Username = username
	err := photoInfo.SetMePhoto(a.DB)

	if err != nil {
		//need logging here instead of print
		return photoInfo
	}

	err2 := photoInfo.GetPhotoInfoByUuid(a.DB)

	if err2 != nil {
		//need logging here instead of print
		fmt.Printf("Something went wrong in db, %v\n", err2)
		return photoInfo
	}

	return photoInfo
}

func (instaData *InstaData) CreateReview(photoInfo db.Photos, a *globalsessionkeeper.AppContext) (*db.Review, error) {

	review := new(db.Review)
	dbRestaurant := new(db.Restaurants)
	//fill in restaurant info
	review.UserID = photoInfo.UserID
	review.Username = photoInfo.Username
	review.Photo.ID = photoInfo.ID
	review.Photo.Uuid = &photoInfo.Uuid
	review.Photo.Latitude = photoInfo.Latitude
	review.Photo.Longitude = photoInfo.Longitude
	review.Restaurant.Name = instaData.Location.Name
	review.Restaurant.Latt = instaData.Location.Latitude
	review.Restaurant.Long = instaData.Location.Longitude
	review.Restaurant.Source = "instagram"
	review.Restaurant.SourceLocID = strconv.FormatInt(instaData.Location.ID, 10)

	// Find Price
	fmt.Printf("\n\n/* //////////////////////////////////////// */\n")
	fmt.Printf("/*                PRICE SEARCH  			*/\n")
	fmt.Printf("/* //////////////////////////////////////// */\n")
	priceRe := regexp.MustCompile(`\$(\d+(\.\d+)?)`)
	price := priceRe.FindStringSubmatch(instaData.Caption.Text)
	fmt.Printf("Price = %v\n", price)
	if len(price) >= 2 {
		if price[1] != "" {
			fmt.Printf("Here = %v\n", price[1])
			f, err :=  strconv.ParseFloat(price[1], 32)
			if err == nil {
				fmt.Printf("Error = NIL\n")
				review.Price = float32(f)
			} else {
				fmt.Printf("convert failed to float\n")
			}
		} else {
			fmt.Printf("Price = Blank")
		}
	}

	dbRestaurant.Name = instaData.Location.Name
	err := dbRestaurant.GetRestaurantInfoByName(a.DB)
	if err != nil && err != sql.ErrNoRows {
		//something bad happened
		fmt.Printf("something went while retrieving data %v\n", err)
		return nil, err
	} else if err == sql.ErrNoRows || dbRestaurant.ID == 0 {
		// not found in DB
		if review.Restaurant.Name != "" {
			fmt.Println("Restaurant Not found in DB, creating new entry")

			err = review.Restaurant.CreateRestaurant(a.DB)
			if err != nil {
				//something bad happened
				fmt.Printf("something went while retrieving data %v", err)
				return nil, err
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
				err = review.Restaurant.CreateRestaurant(a.DB)
				if err != nil {
					//something bad happened
					fmt.Printf("something went while retrieving data %v", err)
					return nil, err
				}
			} else {
				//use existing DB values
				fmt.Println("Location ID Equal, using db values")
				review.Restaurant = *dbRestaurant
			}
		} else if dbRestaurant.Source == "factual"  {
			//trust DB over New
			fmt.Println("Source not same, DB == factual")
			review.Restaurant = *dbRestaurant

		} else if dbRestaurant.Source == "instagram"  && review.Restaurant.Source != "factual" {
			//trust DB over New
			fmt.Println("Source not same, DB == insta")
			review.Restaurant = *dbRestaurant

		} else if review.Restaurant.Source == "instagram" ||
				  review.Restaurant.Source == "factual" {
			fmt.Println("New restaurant instagram or factual, updating db")
			if dbRestaurant.LocationNum == 0 {
				review.Restaurant.UpdateRestaurant(a.DB)
				if err != nil {
					//something bad happened
					fmt.Printf("something went while retrieving data %v", err)
					return nil, err
				}
			} else {
				fmt.Println("location id !=")
				review.Restaurant.LocationNum = dbRestaurant.LocationNum + 1
				err = review.Restaurant.CreateRestaurant(a.DB)
				if err != nil {
					//something bad happened
					fmt.Printf("something went while retrieving data %v", err)
					return nil, err
				}
			}
		}
	}

	for _, tag := range instaData.Tags {
		review.DishTags = append(review.DishTags, db.DishTag{0, tag})
	}

	if instaData.Likes.Count > 0 {
		review.Liked.Bool = true
		review.Liked.Valid = true
	}
	review.Finished.Bool = false
	review.Finished.Valid = true
	review.Description = instaData.Caption.Text

	// create review
	fmt.Printf("Creating reviews.. liked = %v\n", review.Liked)
	return review, nil
}

func AppendIfMissing(slice []int, i int) []int {

    for _, ele := range slice {
        if ele == i {
            return slice
        }
    }
    return append(slice, i)
}

func (googToken *GoogToken) GetToken(w http.ResponseWriter) (*oauth2.Token, error) {

	googTokUrl 	 	:= "https://www.googleapis.com/oauth2/v3/token"
	assert := crypto.CreateJwt(w)
	hc := oauth2.NewClient(context.TODO(), nil)

	if assert.JWT == "" {
		fmt.Println("Couldn't create jwt")
		return &oauth2.Token{}, errors.New("Could not create JWT")
	}

	v := url.Values{}
	v.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	v.Set("assertion", assert.JWT)
	resp, err := hc.PostForm(googTokUrl, v)

	if err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}

	if c := resp.StatusCode; c < 200 || c > 299 {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v\nResponse: %s", resp.Status, body)
	}

	var tokenRes GoogToken
	if err := json.Unmarshal(body, &tokenRes); err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}

	token := &oauth2.Token{
		AccessToken: tokenRes.AccessToken,
		TokenType:   tokenRes.TokenType,
	}

	raw := make(map[string]interface{})
	if err := json.Unmarshal(body, &raw); err == nil {
		token = token.WithExtra(raw)
	}

	if secs := tokenRes.ExpiresIn; secs > 0 {
		token.Expiry = time.Now().Add(time.Duration(secs) * time.Second)
	}

	if v := tokenRes.IDToken; v != "" {
		// decode returned id token to get expiry
		claimSet, err := jws.Decode(v)
		if err != nil {
			return nil, fmt.Errorf("oauth2: error decoding JWT token: %v", err)
		}
		token.Expiry = time.Unix(claimSet.Exp, 0)
	}

	return token, nil
}

func (storageReq *StorageReq) StorePhoto(client *http.Client) error {

	service, err := storage.New(client)
	if err != nil {
		fmt.Printf("Unable to create Storage service: %v\n", err)
		return err
	}

	filename := storageReq.FileName
	bucket := storageReq.Bucket
	fmt.Printf("Bucket = %v, file = %v\n", bucket, filename)

	goFile, err := os.Open(filename)

	if err != nil {

		fmt.Printf("Error opening %v: %v\n", filename, err)
	}

	storageObject, err := service.Objects.Insert(bucket, &storage.Object{Name: storageReq.FileUuid}).Media(goFile).Do()
	fmt.Printf("Got storage.Object, err: %#v, %v", storageObject, err)
	return nil
}

func downloadFile(rawUrl string) (string, error) {

	fmt.Printf("/* //////////////////////////////////////// */\n")
	fmt.Printf("/*                FILE DOWNLOAD 			*/\n")
	fmt.Printf("/* //////////////////////////////////////// */\n")

	fileURL, err := url.Parse(rawUrl)

	if err != nil {
		fmt.Printf("Error = %v\n", err)
		return "", err
	}

	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[4]
	fmt.Printf("Filename = %v\n", fileName)
	file, err := os.Create(FileDownload.FileLoc + fileName)

	if err != nil {
		fmt.Printf("Error = %v\n", err)
		return "", err
	}
	defer file.Close()

	check := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	resp, err := check.Get(rawUrl)

	if err != nil {
		fmt.Printf("Error = %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	fmt.Printf("Status = %v\n", resp.Status)
	size, err := io.Copy(file, resp.Body)

	if err != nil {
		fmt.Printf("Error = %v\n", err)
		return "", err
	}

	fmt.Printf("%s with %v bytes downloaded", fileName, size)
	return fileName, nil
}

func AppCrawl(a *globalsessionkeeper.AppContext, w http.ResponseWriter, r *http.Request) error {

	sessionUser := a.SessionStore.Get("username")
	sessionUserID := a.SessionStore.Get("userId")
	fmt.Println("SessionUser = %v", sessionUser)
	//create variables
	username 	 := reflect.ValueOf(sessionUser).String()
	userId 	 	 := reflect.ValueOf(sessionUserID).Int()

	crawl 	 	 := new(db.Crawl)
	igStore 	 := new(db.IgStore)

	crawl.Username = username
	crawl.UserID = int(userId)
	igStore.UserID = int(userId)
	// var reviews []*db.Reviews

	switch r.Method {

	case "POST":

		/* Initilize Variables*/
		instaData 	:= new(ParentData)

		content, _ := ioutil.ReadAll(r.Body)
		fmt.Printf("Body = %v\n", string(content))

		if err := json.Unmarshal([]byte(content), &instaData); err != nil {
			//need logging here instead of print
			fmt.Printf("something went wrong in app crawl decode %v", err)
			return globalsessionkeeper.ErrorResponse{http.StatusBadRequest, "Malformed JSON: " + err.Error()}
		}

		if len(instaData.Data) == 0 {
			fmt.Println("No New Photos")
			return globalsessionkeeper.ErrorResponse{http.StatusNoContent, "Nothing to update"}
		}

		fmt.Printf("instaData = %v\n", instaData.Data[0])
		fmt.Printf("instaData images = %v\n", instaData.Data[0].Images)
		fmt.Printf("instaData comments = %v\n", instaData.Data[0].Comments)
		fmt.Printf("instaData tags = %v\n", instaData.Data[0].Tags)

		desc, code, reviews, err := DoCrawl(a, username, instaData, false)
		if err != nil {
			fmt.Printf("something went wrong in do crawl %v", err)
			return globalsessionkeeper.ErrorResponse{code, desc}
		}
		/* //////////////////////////////////////// */
		/*               Set Last Crawl 			*/
		/* //////////////////////////////////////// */

		igStore.IgMediaID = instaData.Data[0].ID
		// layOut := "Jan 2, 2006 at 3:04pm (MST)"
		// timeStamp, err := time.Parse(layOut, instaData.Data[0].CreatedTime)
		// // timeStamp, err := time.Parse(instaData.Data[0].CreatedTime, timeStampString)

		// if err != nil {
		// 	fmt.Println(err)
		// 	return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "Not all reviews added: " + err.Error()}
		// }
		// igStore.IgCreatedTime = int(timeStamp.Unix())
		timeEpoch, err := strconv.ParseInt(instaData.Data[0].CreatedTime, 10, 64)
		if err != nil {
			fmt.Println(err)
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "Last Crawl not Saved: " + err.Error()}
		}
		igStore.IgCreatedTime = int(timeEpoch)

		err = igStore.UpdateLastPull(a.DB)
	
		if err != nil {
			fmt.Printf("Could not update table\n")
			return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, "Not all reviews added: " + err.Error()}
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(reviews)
        if err != nil {
        	return globalsessionkeeper.ErrorResponse{http.StatusInternalServerError, err.Error()}
        }
        fmt.Printf("Reviews = %v\n", reviews)
        w.WriteHeader(http.StatusOK)
		return nil

	default:
		return globalsessionkeeper.ErrorResponse{http.StatusMethodNotAllowed, "Method Not Allowed"}
	}
}

func DoCrawl(a *globalsessionkeeper.AppContext, username string, instaData *ParentData, photoUpload bool) (string, int, []*db.Review, error) {

	fmt.Printf("doCrawl: instaData = %v\n", instaData.Data[0])
	fmt.Printf("doCrawl: instaData images = %v\n", instaData.Data[0].Images)
	fmt.Printf("doCrawl: instaData comments = %v\n", instaData.Data[0].Comments)
	fmt.Printf("doCrawl: instaData tags = %v\n", instaData.Data[0].Tags)

	code 		:= 200
	desc 		:= ""
	var err error
	var client *http.Client
	var reviews []*db.Review

	if photoUpload == true {
		client, err = GetGoogleClient()
		if err != nil {
			return "", 500, nil, err
		}
	}

	for elem := 0; elem < len(instaData.Data); elem++ {
		/* //////////////////////////////////////// */
		/*                Create UUID   		    */
		/* //////////////////////////////////////// */
		photoInfo := CreatePhoto(username, a)
		if photoInfo.ID == 0 {
			fmt.Println("Something went wrong to create photo")
			code =  http.StatusPartialContent
			desc = "Not all reviews added: " + err.Error()
		}
		if photoUpload == true {
			/* //////////////////////////////////////// */
			/*                Download File 			*/
			/* //////////////////////////////////////// */
			fileName, err := downloadFile(instaData.Data[elem].Images.StandardRes.Url)
			if err != nil {
				fmt.Printf("No Review Created for %v\n", elem)
				code = http.StatusPartialContent
				desc = "Not all reviews added: " + err.Error()
				continue
			}

			/* //////////////////////////////////////// */
			/*                Store File    		    */
			/* //////////////////////////////////////// */

			storageReq := new(StorageReq)
			storageReq.Bucket = FileDownload.Bucket
			storageReq.FileName = FileDownload.FileLoc + fileName
			storageReq.FileUuid = photoInfo.Uuid
			err = storageReq.StorePhoto(client)
			if err != nil {
				fmt.Println("Something went wrong storing photo")
				code =  http.StatusPartialContent
				desc = "Not all photos added: " + err.Error()
				continue
			}
		}
		/* //////////////////////////////////////// */
		/*                Create Review   		    */
		/* //////////////////////////////////////// */
		review, err1 := instaData.Data[elem].CreateReview(photoInfo, a)
		if photoUpload {
			review.Source = "InstaCrawl"
		} else {
			review.Source = "InstaImport"
		}

		err2 := review.CreateReview(a.DB)

		if err1 == nil && err2 == nil {
			fmt.Printf("Review %v added\n", elem)
			reviews = append(reviews, review)
		} else {
			fmt.Printf("No Review Created for %v\n", elem)
			code =  http.StatusPartialContent
			desc = "Not all reviews added: " + err.Error()
			continue
		}
	}
	return desc, code, reviews, err
}

func GetGoogleClient() (*http.Client, error) {

	googConfig := new(jwt.Config)
	gApiInfo := new(crypto.GApiInfo)

    fileContent, err := ioutil.ReadFile("./chomp_private/Chomp.json")
    if err != nil {
        return new(http.Client), err
    }

    err = json.Unmarshal(fileContent, &gApiInfo)
    if err != nil {
        fmt.Printf("Err = %v", err)
        return new(http.Client), err
    }

    googConfig.Email = gApiInfo.ClientEmail
    googConfig.PrivateKey = []byte(gApiInfo.PrivateKey)
    googConfig.Scopes = []string{`https://www.googleapis.com/auth/devstorage.full_control`}
    googConfig.TokenURL = `https://www.googleapis.com/oauth2/v3/token`

	ctx := context.Background()
	return googConfig.Client(ctx), nil
}
