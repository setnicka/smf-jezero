package server

import (
	"encoding/gob"
	"net/http"
	"os"

	"github.com/gorilla/sessions"

	"github.com/coreos/go-log/log"
)

func (s *Server) isOrg(r *http.Request) bool {
	if org, ok := s.getSession(r).Values["org"].(bool); ok && org {
		return true
	}
	return false
}

func (s *Server) getUser(r *http.Request) string {
	login, _ := s.getSession(r).Values["user"].(string)
	return login
}

func (s *Server) getSession(r *http.Request) *sessions.Session {
	//log := logger.GetDefault()
	session, err := s.sessionStore.Get(r, sessionCookieName)
	if err != nil {
		log.Errorf("Cannot get session '%s': %v", sessionCookieName, err)
		return nil
	}
	return session
}

func (s *Server) checkSession(w http.ResponseWriter, r *http.Request, renew bool) bool {
	session := s.getSession(r)
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

////////////////////////////////////////////////////////////////////////////////

func (s *Server) setFlashMessage(w http.ResponseWriter, r *http.Request, message FlashMessage) {
	// Register the struct so encoding/gob knows about it
	gob.Register(FlashMessage{})

	session, err := s.sessionStore.Get(r, flashSessionName)
	if err != nil {
		return
	}
	session.AddFlash(message)
	err = session.Save(r, w)
	if err != nil {
		log.Errorf("Cannot save flash message: %v", err)
	}
}

func (s *Server) getFlashMessages(w http.ResponseWriter, r *http.Request) []FlashMessage {
	// 1. Get session
	session, err := s.sessionStore.Get(r, flashSessionName)
	if err != nil {
		return nil
	}

	// 2. Get flash messages
	parsedFlashes := []FlashMessage{}
	if flashes := session.Flashes(); len(flashes) > 0 {
		for _, flash := range flashes {
			parsedFlashes = append(parsedFlashes, flash.(FlashMessage))
		}
	}

	// 3. Delete flash messages by saving session
	err = session.Save(r, w)
	if err != nil {
		log.Errorf("Problem during loading flash messages: %v", err)
	}

	return parsedFlashes
}
