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
	"chompapi/crypto"
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
	CreatedTime 	string 	`json:"created_time"`
	Link			string
	Likes 			Likes
	Images 			Images
	Caption 		Caption
	UserHasLiked	bool 	`json:"user_has_liked"`
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
	LowRes		 	Res `json:"low_resolution"`
	Thumbnail		Res `json:"thumbnail"`
	StandardRes 	Res `json:"standard_resolution"`
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
		igStore 	 := new(db.IgStore)

		instaRMediaUrl 	:= "https://api.instagram.com/v1/users/self/media/recent/?access_token=%v&min_timestamp=%v"

		crawl.Username = username
		crawl.UserID = int(userId)
		igStore.UserID = int(userId)

		switch r.Method {

		case "POST":

			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&crawl); err != nil {
				//need logging here instead of print
				fmt.Printf("something went wrong in login %v", err)
				myErrorResponse.Code = http.StatusBadRequest
				myErrorResponse.Error = "Malformed JSON: " + err.Error()
				myErrorResponse.HttpErrorResponder(w)
				return
			}
			fileContent, err := ioutil.ReadFile("./chomp_private/file_download.json")
			if err != nil {
    		    myErrorResponse.Code = http.StatusInternalServerError
    		    myErrorResponse.Error = err.Error()
    		    myErrorResponse.HttpErrorResponder(w)
    		    return
    		}
    		err = json.Unmarshal(fileContent, &FileDownload)
    		if err != nil {
    		    fmt.Printf("Err = %v", err)
    		    myErrorResponse.Code = http.StatusBadRequest
    		    myErrorResponse.Error = "Could not decode"
    		    myErrorResponse.HttpErrorResponder(w)
    		    return
    		}

    		/* //////////////////////////////////////// */
			/*                Check Last Crawl 			*/
			/* //////////////////////////////////////// */

			fmt.Println("=======================================")
			igStore.GetLastPull()
			if err != nil {
				fmt.Printf("Error = %v\n")
				myErrorResponse.Code = http.StatusBadRequest
				myErrorResponse.Error = "Could not set last crawl: " + err.Error()
				myErrorResponse.HttpErrorResponder(w)
			}
			fmt.Printf("\nigStore Pull = %v\n", igStore)
			fmt.Println("=======================================")

			iurl :=  fmt.Sprintf(instaRMediaUrl, crawl.InstaTok, igStore.IgCreatedTime +1)
			request := gorequest.New()
			resp, body, errs := request.Get(iurl).End()

			if errs != nil {
				fmt.Printf("something went wrong in get %v", err)
				myErrorResponse.Code = http.StatusBadRequest
				myErrorResponse.Error = "Malformed JSON: " + err.Error()
				myErrorResponse.HttpErrorResponder(w)
				return
			}

            err = json.Unmarshal([]byte(body), &instaData)

			if err != nil {
				fmt.Printf("Err = %v", err)
				myErrorResponse.Code = http.StatusServiceUnavailable
				myErrorResponse.Error = "Communication Issues:IG: " + err.Error()
				myErrorResponse.HttpErrorResponder(w)
			}

			if len(instaData.Data) == 0 {
				fmt.Println("No New Photos")
				myErrorResponse.Code = http.StatusOK
				myErrorResponse.Error = "Nothing to update"
				myErrorResponse.HttpErrorResponder(w)
				return
			}

			fmt.Printf("Resp:%v \nbody: %v\n, errs: %v\n", resp, body, errs)
			fmt.Printf("instaData = %v\n", instaData.Data[0])
			fmt.Printf("instaData images = %v\n", instaData.Data[0].Images)
			fmt.Printf("instaData comments = %v\n", instaData.Data[0].Comments)
			fmt.Printf("instaData tags = %v\n", instaData.Data[0].Tags)

			var reviewsToWrite []int

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
							reviewsToWrite = AppendIfMissing(reviewsToWrite, index)
						}
					}
				}
			}

			fmt.Printf("\n\n\nreviewsToWrite = %v\n\n\n", reviewsToWrite)
			if len(reviewsToWrite) == 0 {
				fmt.Println("No New Photos")
				myErrorResponse.Code = http.StatusOK
				myErrorResponse.Error = "Nothing to update"
				myErrorResponse.HttpErrorResponder(w)
				return
			}

			if err != nil {
				myErrorResponse.Code = http.StatusServiceUnavailable
				myErrorResponse.Error = "Communication Issues:Google: " + err.Error()
				myErrorResponse.HttpErrorResponder(w)
				return
			}

			googConfig := new(jwt.Config)
			gApiInfo := new(crypto.GApiInfo)
    		fileContent, err = ioutil.ReadFile("./chomp_private/Chomp.json")

    		if err != nil {
    		    myErrorResponse.Code = http.StatusInternalServerError
    		    myErrorResponse.Error = err.Error()
    		    myErrorResponse.HttpErrorResponder(w)
    		    return
    		}
    		err = json.Unmarshal(fileContent, &gApiInfo)
    		if err != nil {
    		    fmt.Printf("Err = %v", err)
    		    myErrorResponse.Code = http.StatusBadRequest
    		    myErrorResponse.Error = "Could not decode"
    		    myErrorResponse.HttpErrorResponder(w)
    		    return
    		}

    		googConfig.Email = gApiInfo.ClientEmail
    		googConfig.PrivateKey = []byte(gApiInfo.PrivateKey)
    		googConfig.Scopes = []string{`https://www.googleapis.com/auth/devstorage.full_control`}
    		googConfig.TokenURL = `https://www.googleapis.com/oauth2/v3/token`

			storageReq := new(StorageReq)
			ctx := context.Background()
			client := googConfig.Client(ctx)

			for i := range reviewsToWrite {

				/* //////////////////////////////////////// */
				/*                Download File 			*/
				/* //////////////////////////////////////// */

				fileName, err := downloadFile(instaData.Data[i].Images.StandardRes.Url)

				if err != nil {
					fmt.Printf("No Review Created for %v\n", i)
					myErrorResponse.Code = http.StatusPartialContent
					myErrorResponse.Error = "Not all reviews added: " + err.Error()
					continue
				}

				/* //////////////////////////////////////// */
				/*                Create UUID   		    */
				/* //////////////////////////////////////// */

				photoInfo := CreatePhoto(username)

				if photoInfo.ID == 0 {
					fmt.Println("Something went wrong to create photo")
					myErrorResponse.Code = http.StatusPartialContent
					myErrorResponse.Error = "Not all reviews added: " + err.Error()
				}

				/* //////////////////////////////////////// */
				/*                Store File    		    */
				/* //////////////////////////////////////// */

				storageReq.Bucket = FileDownload.Bucket
				storageReq.FileName = FileDownload.FileLoc + fileName
				storageReq.FileUuid = photoInfo.Uuid

				err = storageReq.StorePhoto(client)

				if err != nil {
					fmt.Println("Something went wrong storing photo")
					myErrorResponse.Code = http.StatusPartialContent
					myErrorResponse.Error = "Not all photos added: " + err.Error()
					continue
				}

				/* //////////////////////////////////////// */
				/*                Create Review   		    */
				/* //////////////////////////////////////// */

				err = instaData.Data[i].CreateReview(photoInfo)
				if err == nil {

					fmt.Printf("Review %v added\n", i)
				} else {

					fmt.Printf("No Review Created for %v\n", i)
					myErrorResponse.Code = http.StatusPartialContent
					myErrorResponse.Error = "Not all reviews added: " + err.Error()
					continue
				}

				if i == 0 {

					igStore.IgMediaID = instaData.Data[i].ID
					igStore.IgCreatedTime, err = strconv.Atoi(instaData.Data[i].CreatedTime)

					if err != nil {

						fmt.Printf("Cound't convert string%v\n", err)
						myErrorResponse.Code = http.StatusServiceUnavailable
						myErrorResponse.Error = "Communication Issues:IG: " + err.Error()
						myErrorResponse.HttpErrorResponder(w)
					}

				/* //////////////////////////////////////// */
				/*               Set Last Crawl 			*/
				/* //////////////////////////////////////// */

					err = igStore.UpdateLastPull()

					if err != nil {
						fmt.Printf("Could not update table\n")
						myErrorResponse.Code = http.StatusInternalServerError
						myErrorResponse.Error = "Not all reviews added: " + err.Error()
						return
					}
				}
			}

		default:

			myErrorResponse.Code = http.StatusMethodNotAllowed
			myErrorResponse.Error = "Method Not Allowed"
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
		return photoInfo
	} 

	err2 := photoInfo.GetPhotoInfoByUuid()

	if err2 != nil {
		//need logging here instead of print
		fmt.Printf("Something went wrong in db, %v\n", err2)
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
	review.Photo.ID = photoInfo.ID
	review.Restaurant.Name = instaData.Location.Name
	review.Restaurant.Latt = instaData.Location.Latitude
	review.Restaurant.Long = instaData.Location.Longitude
	review.Restaurant.Source = "instagram"
	review.Restaurant.SourceLocID = strconv.FormatInt(instaData.Location.ID, 10)

	dbRestaurant.Name = instaData.Location.Name
	err := dbRestaurant.GetRestaurantInfoByName()
	if err != nil && err != sql.ErrNoRows{
		//something bad happened
		fmt.Printf("something went while retrieving data %v\n", err)
		return err
	} else if err == sql.ErrNoRows || dbRestaurant.ID == 0 {
		// not found in DB
		if review.Restaurant.Name != "" {
			fmt.Println("Restaurant Not found in DB, creating new entry")

			err = review.Restaurant.CreateRestaurant()
			if err != nil {
				//something bad happened
				fmt.Printf("something went while retrieving data %v", err)
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

		} else if review.Restaurant.Source == "instagram" {
			fmt.Println("New restaurant instagram, updating db")
			if dbRestaurant.LocationNum == 0 {
				review.Restaurant.UpdateRestaurant()
				if err != nil {
					//something bad happened
					fmt.Printf("something went while retrieving data %v", err)
					return err
				}
			} else {
				fmt.Println("location id !=")
				review.Restaurant.LocationNum = dbRestaurant.LocationNum + 1
				err = review.Restaurant.CreateRestaurant()
				if err != nil {
					//something bad happened
					fmt.Printf("something went while retrieving data %v", err)
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
		review.Liked.Valid = true
	}
	review.Finished.Bool = false
	review.Finished.Valid = true
	review.Description = instaData.Caption.Text

	// create review
	fmt.Printf("Creating reviews.. liked = %v\n", review.Liked)
	err = review.CreateReview()
	if err != nil {
		//something bad happened
		fmt.Printf("something went while retrieving data %v", err)
		return err
	}
	return nil
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
	json.Unmarshal(body, &raw) // no error checks for optional fields
	token = token.WithExtra(raw)

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


