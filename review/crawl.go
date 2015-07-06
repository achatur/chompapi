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
    // "google.golang.org/api/googleapi/transport"
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
			fmt.Println("=======================================")
			// igStore.GetLastPull()
			// if err != nil {
			// 	fmt.Printf("Error = %v\n")
			// }
			// fmt.Printf("\nigStore Pull = %v\n", igStore)
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
			//newTS := new(oauth2.TokenSource)
			googToken := new(GoogToken)
			//var tokenSource oauth2.TokenSource
			//token, err := googToken.GetToken(w)
			//tokenSource = googToken.GetToken(w)
			if err != nil {
				myErrorResponse.Code = http.StatusServiceUnavailable
				myErrorResponse.Error = "Communication Issues:Google: " + err.Error()
				myErrorResponse.HttpErrorResponder(w)
				return
			}
			// fmt.Printf("\n\n========\nNewTS = %v\n=======\n", token)
			// privateKey, err := ioutil.ReadFile("./chomp_private/Chomp.pem")
			// //privateKey, err := ioutil.ReadFile("./chomp_private/Chomp.p12")
			// if err != nil {
   //  		    myErrorResponse.Code = http.StatusInternalServerError
   //  		    myErrorResponse.Error = err.Error()
   //  		    myErrorResponse.HttpErrorResponder(w)
   //  		    return
   //  		}
			googConfig := new(jwt.Config)
			gApiInfo := new(crypto.GApiInfo)
			// googConfig.Email = "486543155383-oo5gldbn5q9jm3mei3de3p5p95ffn8fi@developer.gserviceaccount.com"
			// //googConfig.Email = "486543155383-oo5gldbn5q9jm3mei3de3p5p95ffn8fi.apps.googleusercontent.com"
			// //googConfig.Subject = "486543155383-oo5gldbn5q9jm3mei3de3p5p95ffn8fi.apps.googleusercontent.com"
			// googConfig.PrivateKey = privateKey
			// // googConfig.Scopes =  append(googConfig.Scopes, "https://www.googleapis.com/auth/devstorage.full_control")
			// googConfig.Scopes =  []string{"https://www.googleapis.com/auth/devstorage.full_control"}
			// googConfig.TokenURL = "https://www.googleapis.com/oauth2/v3/token"
    		//gApiInfo := new(GApiInfo)
    		fileContent, err = ioutil.ReadFile("./chomp_private/Chomp.json")
    		//privateKey, err := ioutil.ReadFile("./chomp_private/Chomp.pem")
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
    		fmt.Printf("Json = %v\n", gApiInfo)
    		// Set some claims
    		// token.Claims["scope"] = `https://www.googleapis.com/auth/devstorage.full_control`
    		// // token.Claims["iss"] = gApiInfo.ClientId
    		// token.Claims["iss"] = gApiInfo.ClientEmail
    		// token.Claims["iat"] = time.Now().Unix()
    		// token.Claims["exp"] = time.Now().Add(time.Hour * 1).Unix()
    		// token.Claims["aud"] = `https://www.googleapis.com/oauth2/v3/token`
    		// fmt.Printf("Token Claims: %v\n", token.Claims)
    		googConfig.Email = gApiInfo.ClientEmail
    		googConfig.PrivateKey = []byte(gApiInfo.PrivateKey)
    		googConfig.Scopes = []string{`https://www.googleapis.com/auth/devstorage.full_control`}
    		googConfig.TokenURL = `https://www.googleapis.com/oauth2/v3/token`

			storageReq := new(StorageReq)
			// storageReq.Token = *token
			ctx := context.Background()
			// client := googConfig.Client(ctx) 
			//client := oauth2.NewClient(context.TODO(), nil)
			token, err := googToken.GetToken(w)
			storageReq.Token = token
			fmt.Printf("Token = %v\n", token)
			//ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{ Transport: &transport.APIKey{ Key: token.AccessToken } })
			//client := oauth2.NewClient(context.Background(), googConfig.TokenSource(context.TODO()))
			fmt.Printf("creating client with %v\n", googConfig)
			client := googConfig.Client(ctx)

			// tokenSource := googConfig.TokenSource(ctx)
			// //fmt.Printf("TokenSource = %v\n", tokenSource)
			// client = oauth2.NewClient(ctx, tokenSource)
			// token2, err := tokenSource.Token()
			// fmt.Printf("Token = %v, err = %v\n", token2, err)

			// newTS := new oauth2.TokenSource


			for i := range reviewsToWrite {

				/*====================================//
				// generate and store UUID for photo  //
				// upload file to google storeage     //
				// update database                    //
				//====================================*/

				photoInfo := CreatePhoto(username)
				if photoInfo.ID == 0 {
					fmt.Println("Something went wrong to create photo")
					myErrorResponse.Code = http.StatusPartialContent
					myErrorResponse.Error = "Not all reviews added: " + err.Error()
				}

				err := instaData.Data[i].CreateReview(photoInfo)
				if err == nil {

					fmt.Printf("Review %v added\n", i)
				} else {

					fmt.Printf("No Review Created for %v\n", i)
					myErrorResponse.Code = http.StatusPartialContent
					myErrorResponse.Error = "Not all reviews added: " + err.Error()
				}
				/* //////////////////////////////////////// */
				/*                FILE DOWNLOAD 			*/
				/* //////////////////////////////////////// */

				fileName, err := downloadFile(instaData.Data[i].Images.StandardRes.Url)
				if err != nil {
					fmt.Printf("No Review Created for %v\n", i)
					myErrorResponse.Code = http.StatusPartialContent
					myErrorResponse.Error = "Not all reviews added: " + err.Error()
				}

				/* //////////////////////////////////////// */
				/*                FILE DOWNLOAD END			*/
				/* //////////////////////////////////////// */

				storageReq.Bucket = "chomp_photos"
				//storageReq.FileName = "/home/amir.chatur/file_serve/game_chair_box.jpg"
				storageReq.FileName = FileDownload.FileLoc + fileName
				storageReq.FileUuid = photoInfo.Uuid
				err = storageReq.StorePhoto(client)
				if err != nil {
					fmt.Println("Something went wrong storing photo")
					myErrorResponse.Code = http.StatusPartialContent
					myErrorResponse.Error = "Not all photos added: " + err.Error()
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

					// err = igStore.UpdateLastPull()

					// if err != nil {
					// 	fmt.Printf("Could not update table\n")
					// 	myErrorResponse.Code = http.StatusInternalServerError
					// 	myErrorResponse.Error = "Not all reviews added: " + err.Error()
					// 	return
					// }
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
		// review.Liked.Value = true
	}
	review.Finished.Bool = false
	review.Description = instaData.Caption.Text

	// create review
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
	// googGT 	 	 	:= "grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Ajwt-bearer&assertion="
	assert := crypto.CreateJwt(w)
	// request := gorequest.New()
	//token := new(oauth2.Token)
	hc := oauth2.NewClient(context.TODO(), nil)

	if assert.JWT == "" {
		fmt.Println("Couldn't create jwt")
		//return token, errors.New("Could not create JWT")
		return &oauth2.Token{}, errors.New("Could not create JWT")
	}
	fmt.Printf("====\n")
	fmt.Printf("Assert = %v\n", assert.JWT)
	fmt.Printf("====\n")
	//url :=  fmt.Sprintf(googTokUrl, googGT + assert.JWT)
	// resp, body, errs := request.Post(url).End()
	v := url.Values{}
	v.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	v.Set("assertion", assert.JWT)
	//v.Set("assertion", "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiI0ODY1NDMxNTUzODMtb281Z2xkYm41cTlqbTNtZWkzZGUzcDVwOTVmZm44ZmlAZGV2ZWxvcGVyLmdzZXJ2aWNlYWNjb3VudC5jb20iLCJzY29wZSI6Imh0dHBzOi8vd3d3Lmdvb2dsZWFwaXMuY29tL2F1dGgvZGV2c3RvcmFnZS5mdWxsX2NvbnRyb2wiLCJhdWQiOiJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9vYXV0aDIvdjMvdG9rZW4iLCJleHAiOjE0MzYxMTgxMzAsImlhdCI6MTQzNjExNDUzMH0.H689BC46sOoGm3Nu_OC5EcQq3Fr9iRG3wlCORlsc5k0HZl_gqbyE3t5HRw2uDkyah_uAjs-nOFj_feUuDyF_dtzx_4AcqKbBP2xn45f5OZlRMHgfDnLZeZoJY0_WK0G6rixPSl6SP6FFdlQ7lRVJPtH_-zZzCxMRHoD5YJAW7_wnWfs3jgCvYMzwuwSwZxe7gDeFoeZlpB7aNAAQqwZJlEdXutZmWQUV6wGOSMoa__ngYBc0JaPDIzLoi8AiEgZY9bBZONs9OhcDLOVf5CvRH-9nnhaBprEHzkfqjTWK3x51taNyODithiSpR6ENheafikFFHWUd2DijYQFWzQqvfg")
	//v.Set("assertion", "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9vYXV0aDIvdjMvdG9rZW4iLCJleHAiOjE0MzYxMjA0NjAsImlhdCI6MTQzNjExNjg2MCwiaXNzIjoiNDg2NTQzMTU1MzgzLTR2cXVhcW81bTBpYms1Y24ybm8xaTk4azVsdjA5ZHBzQGRldmVsb3Blci5nc2VydmljZWFjY291bnQuY29tIiwic2NvcGUiOiJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9hdXRoL2RldnN0b3JhZ2UuZnVsbF9jb250cm9sIn0.cG3U8peemtCzPOBFraa7Q9DlP1H-XecocQmxY0XEnTDIt8UnGQnUpMU4HTdAob_9X569oBMwY8ZAlsyy2tGBuwIT9UlRt6N2DJ4jc7PcggtGGblfngczSw6j5-4af-T9CDZscNBHoNsrVm_WV_0D2D1--YC2A0eikQGOutSVJQv4Na-EOB0Orf-sSGbxxlCq0UrdBtVa3fKnVnnlwrGwnPw1S-Vndxfb8UYadG8fu1pxWW9Jbx8knkXXr86ZZSri3CjZuSVpbHY4iVIjn35GmqZKai7uoeNpUqOauaImnHLglgrBsmOMkLwmTQkOxCpMGvZsxsINoURyj9Z3wUavZg")
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
	// // tokenRes is the JSON response body.
	// var tokenRes struct {
	// 	AccessToken string `json:"access_token"`
	// 	TokenType   string `json:"token_type"`
	// 	IDToken     string `json:"id_token"`
	// 	ExpiresIn   int64  `json:"expires_in"` // relative seconds from now
	// }
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

	// if errs != nil {
	// 	fmt.Printf("something went wrong in get %v", errs)
	// 	// myErrorResponse.Code = http.StatusServiceUnavailable
	// 	// myErrorResponse.Error = "Communication Issues:Google: " + errs.Error()
	// 	// myErrorResponse.HttpErrorResponder(w)
	// 	return token, errs[0]
	// }
	// fmt.Printf("\n\ninside getToken\nResp:%v \nbody: %v\n, errs: %v\n\n\n", resp, body, errs)

	// err := json.Unmarshal([]byte(body), &googToken)
	// //err := json.Unmarshal([]byte(body), &token)


	// if err != nil {
	// 	fmt.Printf("Err = %v", err)
	// 	return token, err
	// }
	// fmt.Printf("GetToken jwt = %v\n", token)
	return token, nil
}

func (storageReq *StorageReq) StorePhoto(client *http.Client) error {
	
	// if len(argv) != 2 {
	// 	fmt.Fprintln(os.Stderr, "Usage: storage filename bucket (to upload an object)")
	// 	return
	// }
	fmt.Printf("Client = %v\n", client.Transport)
	// client.Set("Authorization", storageReq.Token.AccessToken)

	service, err := storage.New(client)
	if err != nil {
		// log.Fatalf("Unable to create Storage service: %v", err)
		fmt.Printf("Unable to create Storage service: %v\n", err)
		return err
	}
	// service.Channel.Token = storageReq.Token
	filename := storageReq.FileName
	bucket := storageReq.Bucket

	fmt.Printf("Bucket = %v, file = %v\n", bucket, filename)

	goFile, err := os.Open(filename)
	if err != nil {
		// log.Fatalf("error opening %q: %v", filename, err)
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
	//fileURL, err := url.Parse(instaData.Data[i].Images.StandardRes.Url)
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


