// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// package session provider
//
// Usage:
// import(
//   "github.com/achatur/beego/session"
// )
//
//	func init() {
//      globalSessions, _ = session.NewManager("memory", `{"cookieName":"gosessionid", "enableSetCookie,omitempty": true, "gclifetime":3600, "maxLifetime": 3600, "secure": false, "sessionIDHashFunc": "sha1", "sessionIDHashKey": "", "cookieLifeTime": 3600, "providerConfig": ""}`)
//		go globalSessions.GC()
//	}
//
// more docs: http://beego.me/docs/module/session.md
package session

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SessionStore contains all data for one session process with specific id.
type SessionStore interface {
	Set(key, value interface{}) error     //set session value
	Get(key interface{}) interface{}      //get session value
	Delete(key interface{}) error         //delete session value
	SessionID() string                    //back current sessionID
	SessionRelease(w http.ResponseWriter) // release the resource & save data to provider & return the data
	Flush() error                         //delete all data
}

// Provider contains global session methods and saved SessionStores.
// it can operate a SessionStore by its id.
type Provider interface {
	SessionInit(gclifetime int64, config string) error
	SessionRead(sid string) (SessionStore, error)
	SessionExist(sid string) bool
	SessionRegenerate(oldsid, sid string) (SessionStore, error)
	SessionDestroy(sid string) error
	SessionAll() int //get all active session
	SessionGC()
}

var provides = make(map[string]Provider)

// Register makes a session provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, provide Provider) {
	if provide == nil {
		panic("session: Register provide is nil")
	}
	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provider " + name)
	}
	provides[name] = provide
}

type managerConfig struct {
	CookieName      string `json:"cookieName"`
	EnableSetCookie bool   `json:"enableSetCookie,omitempty"`
	Gclifetime      int64  `json:"gclifetime"`
	Maxlifetime     int64  `json:"maxLifetime"`
	Secure          bool   `json:"secure"`
	CookieLifeTime  int    `json:"cookieLifeTime"`
	ProviderConfig  string `json:"providerConfig"`
	Domain          string `json:"domain"`
	SessionIdLength int64  `json:"sessionIdLength"`
}

// Manager contains Provider and its configuration.
type Manager struct {
	provider Provider
	config   *managerConfig
}

// Create new Manager with provider name and json config string.
// provider name:
// 1. cookie
// 2. file
// 3. memory
// 4. redis
// 5. mysql
// json config:
// 1. is https  default false
// 2. hashfunc  default sha1
// 3. hashkey default beegosessionkey
// 4. maxage default is none
func NewManager(provideName, config string) (*Manager, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}
	cf := new(managerConfig)
	cf.EnableSetCookie = true
	err := json.Unmarshal([]byte(config), cf)
	if err != nil {
		return nil, err
	}
	if cf.Maxlifetime == 0 {
		cf.Maxlifetime = cf.Gclifetime
	}
	err = provider.SessionInit(cf.Maxlifetime, cf.ProviderConfig)
	if err != nil {
		return nil, err
	}

	if cf.SessionIdLength == 0 {
		cf.SessionIdLength = 16
	}

	return &Manager{
		provider,
		cf,
	}, nil
}

// Start session. generate or read the session id from http request.
// if session id exists, return SessionStore with this id.
func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (session SessionStore, err error) {
	fmt.Printf("in SessionStart\n")
	cookie, errs := r.Cookie(manager.config.CookieName)
	if errs != nil || cookie.Value == "" {
		fmt.Printf("Cookies blank\n")
		sid, errs := manager.sessionId(r)
		if errs != nil {
			return nil, errs
		}
		session, err = manager.provider.SessionRead(sid)
		fmt.Printf("SessionStart Session = %v\n", session)
		fmt.Printf("SessionStart cookie config = %v\n", manager.config)
		cookie = &http.Cookie{Name: manager.config.CookieName,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
			Secure:   manager.config.Secure,
			Domain:   manager.config.Domain}
		if manager.config.CookieLifeTime >= 0 {
			fmt.Printf("CookieLifeTime = %v\n", manager.config.CookieLifeTime)
			cookie.MaxAge = manager.config.CookieLifeTime
		}
		if manager.config.EnableSetCookie {
			fmt.Printf("SessionStore enableSetCookie= %v\n", manager.config.EnableSetCookie)
			http.SetCookie(w, cookie)
		}
		r.AddCookie(cookie)
	} else {
		fmt.Printf("Cookie not nil or blank, %v\n", cookie.Value)
		sid, errs := url.QueryUnescape(cookie.Value)
		if errs != nil {
			return nil, errs
		}
		if manager.provider.SessionExist(sid) {
			fmt.Printf("SessionStore sessionExists\n")
			session, err = manager.provider.SessionRead(sid)
			fmt.Printf("SessionStore session %v\n", session)
		} else {
			sid, err = manager.sessionId(r)
			fmt.Printf("SessionStore dne, sid %v\n", sid)
			if err != nil {
				return nil, err
			}
			session, err = manager.provider.SessionRead(sid)
			fmt.Printf("SessionStore dne session: %v\n", session)
			cookie = &http.Cookie{Name: manager.config.CookieName,
				Value:    url.QueryEscape(sid),
				Path:     "/",
				HttpOnly: true,
				Secure:   manager.config.Secure,
				Domain:   manager.config.Domain}
			if manager.config.CookieLifeTime >= 0 {
				cookie.MaxAge = manager.config.CookieLifeTime
			}
			if manager.config.EnableSetCookie {
				fmt.Printf("SessionStore dne cookie: %v\n", cookie)
				http.SetCookie(w, cookie)
			}
			r.AddCookie(cookie)
		}
	}
	return
}

// Destroy session by its id in http request cookie.
func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.config.CookieName)
	fmt.Printf("SessionDestroy: cookie = %v\n", cookie)
	if err != nil || cookie.Value == "" {
		return
	} else {
		manager.provider.SessionDestroy(cookie.Value)
		expiration := time.Now()
		cookie := http.Cookie{Name: manager.config.CookieName,
			Path:     "/",
			HttpOnly: true,
			Expires:  expiration,
			MaxAge:   -1}
		http.SetCookie(w, &cookie)
	}
}

// Get SessionStore by its id.
func (manager *Manager) GetSessionStore(sid string) (sessions SessionStore, err error) {
	sessions, err = manager.provider.SessionRead(sid)
	return
}

// Start session gc process.
// it can do gc in times after gc lifetime.
func (manager *Manager) GC() {
	manager.provider.SessionGC()
	time.AfterFunc(time.Duration(manager.config.Gclifetime)*time.Second, func() { manager.GC() })
}

// Regenerate a session id for this SessionStore who's id is saving in http request.
func (manager *Manager) SessionRegenerateId(w http.ResponseWriter, r *http.Request) (session SessionStore) {
	sid, err := manager.sessionId(r)
	if err != nil {
		return
	}
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil && cookie.Value == "" {
		//delete old cookie
		session, _ = manager.provider.SessionRead(sid)
		cookie = &http.Cookie{Name: manager.config.CookieName,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
			Secure:   manager.config.Secure,
			Domain:   manager.config.Domain,
		}
	} else {
		oldsid, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.provider.SessionRegenerate(oldsid, sid)
		cookie.Value = url.QueryEscape(sid)
		cookie.HttpOnly = true
		cookie.Path = "/"
	}
	if manager.config.CookieLifeTime >= 0 {
		cookie.MaxAge = manager.config.CookieLifeTime
	}
	http.SetCookie(w, cookie)
	r.AddCookie(cookie)
	return
}

// Get all active sessions count number.
func (manager *Manager) GetActiveSession() int {
	return manager.provider.SessionAll()
}

// Set cookie with https.
func (manager *Manager) SetSecure(secure bool) {
	manager.config.Secure = secure
}

func (manager *Manager) sessionId(r *http.Request) (string, error) {
	b := make([]byte, manager.config.SessionIdLength)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		return "", fmt.Errorf("Could not successfully read from the system CSPRNG.")
	}
	return hex.EncodeToString(b), nil
}
