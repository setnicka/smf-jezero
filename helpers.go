package main

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"

	"github.com/coreos/go-log/log"
)

func isOrg(r *http.Request) bool {
	if org, ok := getSession(r).Values["org"].(bool); ok && org {
		return true
	}
	return false
}

func getUser(r *http.Request) string {
	login, _ := getSession(r).Values["user"].(string)
	return login
}

func getSession(r *http.Request) *sessions.Session {
	//log := logger.GetDefault()
	session, err := server.sessionStore.Get(r, SESSION_COOKIE_NAME)
	if err != nil {
		log.Errorf("Cannot get session '%s': %v", SESSION_COOKIE_NAME, err)
		return nil
	}
	return session
}

func checkSession(w http.ResponseWriter, r *http.Request, renew bool) bool {
	session := getSession(r)
	if session == nil {
		return false
	}
	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		return false
	}

	if renew {
		session.Save(r, w)
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////

type noListFile struct {
	http.File
}

func (f noListFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

// NoListFileSystem is used for accessing static resources but without listing directory index
type NoListFileSystem struct {
	base http.FileSystem
}

// Open opens dir/file on given path
func (fs NoListFileSystem) Open(name string) (http.File, error) {
	f, err := fs.base.Open(name)
	if err != nil {
		return nil, err
	}
	return noListFile{f}, nil
}
