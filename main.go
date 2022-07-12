package main

import (
	"flag"
	"html/template"
	"net/http"
	"sync"
	"time"

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

var (
	listen = flag.String("listen", ":8080", "Listen address")
)

type Server struct {
	sessionStore      sessions.Store
	templates         *template.Template
	state             *game.State
	countdownDuration time.Duration
	countdownTo       time.Time
	countdownTimer    *time.Timer
	mutex             sync.RWMutex
}

// Global singleton
var server *Server

////////////////////////////////////////////////////////////////////////////////

func main() {
	cookieStore := sessions.NewCookieStore([]byte(SESSION_SECRET))
	cookieStore.MaxAge(SESSION_MAX_AGE)
	//cookieStore.Options.Domain = ".fuf.me"

	server = &Server{
		sessionStore:   cookieStore,
		state:          game.Init(),
		countdownTimer: time.NewTimer(time.Hour),
	}
	server.countdownTimer.Stop() // by default stop the timer, but we need to initialize it

	server.Start()
}

////////////////////////////////////////////////////////////////////////////////

func (s *Server) Start() {
	log.Info("Starting server...")
	flag.Parse()

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
	router.HandleFunc("/org/charts", authOrg(orgChartsHandler))
	router.HandleFunc("/org/getHash", authOrg(orgHashHandler))

	// Teams handlers
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/", auth(s, teamIndexHandler))
	router.HandleFunc("/getHash", auth(s, teamHashHandler))

	// Dashboard
	router.HandleFunc("/view", viewIndexHandler)
	router.HandleFunc("/view/status", viewStatusHandler)
	router.HandleFunc("/view/getHash", viewHashHandler)

	// 2. Load templates
	if _, err := s.getTemplates(); err != nil {
		log.Errorf("Cannot load templates: %v", err)
		return
	}

	// 3. Start countdown timer
	go func() {
		for range s.countdownTimer.C {
			s.mutex.Lock()
			log.Infof("Next round by timer")
			server.state.EndRound()
			s.resetTimer()
			s.mutex.Unlock()
		}
	}()

	// 4. Listen on given port
	log.Infof("Server started on %s", *listen)
	err := http.ListenAndServe(*listen, router)
	if err != nil {
		log.Errorf("ERROR: %v", err)
	}
}

func (s *Server) stopTimer() {
	server.countdownTo = time.Time{}
	s.countdownTimer.Stop()
}

func (s *Server) resetTimer() {
	if s.countdownDuration == 0 {
		return
	}
	s.stopTimer()
	s.countdownTo = time.Now().Add(s.countdownDuration)
	s.countdownTimer.Reset(s.countdownDuration)
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
