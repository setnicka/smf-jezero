package main

import (
	"html/template"
	"net/http"

	"github.com/coreos/go-log/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/setnicka/smf-jezero/game"
)

const (
	SESSION_SECRET      = "bojovyVlkodlakCeskePosty"
	SESSION_MAX_AGE     = 3600 * 24
	SESSION_COOKIE_NAME = "cp_cookie"
	TEMPLATE_DIR        = "templates"
	STATIC_DIR          = "static"

	ORG_LOGIN    = "smf"
	ORG_PASSWORD = "tragedieobecnipastviny" // TODO: load from config file?
)

type Server struct {
	sessionStore sessions.Store
	templates    *template.Template
	state        *game.State
}

// Global singleton
var server *Server

////////////////////////////////////////////////////////////////////////////////

func main() {
	cookieStore := sessions.NewCookieStore([]byte(SESSION_SECRET))
	cookieStore.MaxAge(SESSION_MAX_AGE)
	//cookieStore.Options.Domain = ".fuf.me"

	server = &Server{
		sessionStore: cookieStore,
		state:        game.Init(),
	}

	server.Start()
}

////////////////////////////////////////////////////////////////////////////////

func (s *Server) Start() {
	log.Info("Starting server...")

	// 1. Construct router
	router := mux.NewRouter()

	// Static resources
	fs := NoListFileSystem{http.Dir(STATIC_DIR)}
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(fs)))

	// Org handlers
	router.HandleFunc("/org", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/org/teams", http.StatusSeeOther)
	})
	router.HandleFunc("/org/login", orgLoginHandler)
	router.HandleFunc("/org/teams", authOrg(orgTeamsHandler))
	router.HandleFunc("/org/dashboard", authOrg(orgDashboardHandler))

	// Teams handlers
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/", auth(s, teamIndexHandler))

	router.HandleFunc("/getRound", getRoundHandler)

	// 2. Load templates
	if _, err := s.getTemplates(); err != nil {
		log.Errorf("Cannot load templates: %v", err)
		return
	}

	// 3. Listen on given port
	log.Info("Server started")
	http.ListenAndServe(":8080", router)
}

func auth(server *Server, handle http.HandlerFunc, renewAuth ...bool) http.HandlerFunc {
	renew := true
	if len(renewAuth) > 0 {
		renew = renewAuth[0]
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if checkSession(w, r, renew) {
			if server.state.GetTeam(getUser(r)) != nil {
				handle(w, r)
				return
			}
		}
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}
}

func authOrg(handle http.HandlerFunc, renewAuth ...bool) http.HandlerFunc {
	renew := true
	if len(renewAuth) > 0 {
		renew = renewAuth[0]
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if checkSession(w, r, renew) && isOrg(r) {
			handle(w, r)
			return
		}
		http.Redirect(w, r, "/org/login", http.StatusTemporaryRedirect)
	}
}
