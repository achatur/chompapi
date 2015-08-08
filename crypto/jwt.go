package crypto

import (
    "cmd/github.com/dgrijalva/jwt-go"
    "encoding/json"
    "net/http"
    "cmd/chompapi/globalsessionkeeper"
    "io/ioutil"
    "fmt"
    "time"
)

type JWT struct {
  JWT string `json:"jwt"`
}

type GApiInfo struct {
    PrivateKeyId    string `json:"private_key_id"`
    PrivateKey      string `json:"private_key"`
    ClientEmail     string `json:"client_email"`
    ClientId        string `json:"client_id"`
    Type            string `json:"type"`
}

var MyErrorResponse globalsessionkeeper.ErrorResponse

func GetJwt(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    jwt := CreateJwt(w)
    if jwt.JWT == "" {
        fmt.Printf("Empty..jwt%v \n", jwt.JWT)
        return
    }
    json.NewEncoder(w).Encode(jwt)
    return
}

func CreateJwt(w http.ResponseWriter) JWT {
    token := jwt.New(jwt.SigningMethodRS256)
    gApiInfo := new(GApiInfo)
    fileContent, err := ioutil.ReadFile("./chomp_private/Chomp.json")
    privateKey, err := ioutil.ReadFile("./chomp_private/Chomp.pem")
    if err != nil {
        MyErrorResponse.Code = http.StatusInternalServerError
        MyErrorResponse.Error = err.Error()
        MyErrorResponse.HttpErrorResponder(w)
        return JWT{}
    }
    err = json.Unmarshal(fileContent, &gApiInfo)
    if err != nil {
        fmt.Printf("Err = %v", err)
        MyErrorResponse.Code = http.StatusBadRequest
        MyErrorResponse.Error = "Could not decode"
        MyErrorResponse.HttpErrorResponder(w)
        return JWT{}
    }
    fmt.Printf("Json = %v\n", gApiInfo)
    // Set some claims
    token.Claims["scope"] = `https://www.googleapis.com/auth/devstorage.full_control`
    //token.Claims["iss"] = gApiInfo.ClientId
    token.Claims["iss"] = gApiInfo.ClientEmail
    token.Claims["iat"] = time.Now().Unix()
    token.Claims["exp"] = time.Now().Add(time.Hour * 1).Unix()
    token.Claims["aud"] = `https://www.googleapis.com/oauth2/v3/token`
    fmt.Printf("Token Claims: %v\n", token.Claims)
    // Sign and get the complete encoded token as a string
    tokenString, err := token.SignedString(privateKey)
    if err != nil {
        fmt.Printf("Err = %v\n", err)
        MyErrorResponse.Code = 500
        MyErrorResponse.Error = err.Error()
        return JWT{}
    }
    fmt.Printf("tokenString = %v\n", tokenString)
    jwt := JWT{tokenString}
    fmt.Printf("Jwt = %v\n", jwt)
    return jwt
}
