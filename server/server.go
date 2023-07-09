package server

import (
	"flag"
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/setnicka/smf-jezero/config"
	"github.com/setnicka/smf-jezero/game"
)

// Server for the game
type Server struct {
	cfg               config.ServerConfig
	sessionStore      sessions.Store
	templates         *template.Template
	state             *game.State
	variant           game.Variant
	countdownDuration time.Duration
	nextCountdown     time.Duration
	countdownTo       time.Time
	countdownTimer    *time.Timer
	mutex             sync.RWMutex
}

// New creates new server
func New(cfg config.ServerConfig, game *game.State, variant game.Variant) *Server {
	cookieStore := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	cookieStore.MaxAge(cfg.SessionMaxAge)
	//cookieStore.Options.Domain = ".fuf.me"

	s := &Server{
		cfg:            cfg,
		sessionStore:   cookieStore,
		state:          game,
		variant:        variant,
		countdownTimer: time.NewTimer(time.Hour),
	}
	s.countdownTimer.Stop() // by default stop the timer, but we need to initialize it

	return s
}

// Start the HTTP server for the game
func (s *Server) Start() {
	log.Info("Starting server...")
	flag.Parse()

	// 1. Construct router
	router := mux.NewRouter().StrictSlash(true)

	// Static resources
	fs := NoListFileSystem{http.Dir(s.cfg.StaticDir)}
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(fs)))

	auth := func(handle http.HandlerFunc, renewAuth ...bool) http.HandlerFunc {
		renew := true
		if len(renewAuth) > 0 {
			renew = renewAuth[0]
		}

		return func(w http.ResponseWriter, r *http.Request) {
			if s.checkSession(w, r, renew) {
				if s.state.GetTeam(s.getUser(r)) != nil {
					handle(w, r)
					return
				}
			}
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		}
	}

	authOrg := func(handle http.HandlerFunc, renewAuth ...bool) http.HandlerFunc {
		renew := true
		if len(renewAuth) > 0 {
			renew = renewAuth[0]
		}

		return func(w http.ResponseWriter, r *http.Request) {
			if s.checkSession(w, r, renew) && s.isOrg(r) {
				handle(w, r)
				return
			}
			http.Redirect(w, r, "/org/login", http.StatusTemporaryRedirect)
		}
	}

	// Org handlers
	router.HandleFunc("/org", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/org/teams", http.StatusSeeOther)
	})
	router.HandleFunc("/org/login", s.orgLoginHandler)
	router.HandleFunc("/org/teams", authOrg(s.orgTeamsHandler))
	router.HandleFunc("/org/dashboard", authOrg(s.orgDashboardHandler))
	router.HandleFunc("/org/dashboard/table", authOrg(s.orgDashboardTableHandler))
	router.HandleFunc("/org/charts", authOrg(s.orgChartsHandler))
	router.HandleFunc("/org/getHash", authOrg(s.orgHashHandler))

	// Teams handlers
	router.HandleFunc("/login", s.loginHandler)
	router.HandleFunc("/", auth(s.teamIndexHandler))
	router.HandleFunc("/getHash", auth(s.teamHashHandler))

	// Dashboard
	router.HandleFunc("/view", s.viewIndexHandler)
	router.HandleFunc("/view/status", s.viewStatusHandler)
	router.HandleFunc("/view/getHash", s.viewHashHandler)

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
			s.state.EndRound()
			s.countdownDuration = s.nextCountdown
			s.resetTimer()
			s.mutex.Unlock()
		}
	}()

	// 4. Listen on given port
	log.Infof("Server started on %s", s.cfg.Listen)
	err := http.ListenAndServe(s.cfg.Listen, router)
	if err != nil {
		log.Errorf("ERROR: %v", err)
	}
}

func (s *Server) stopTimer() {
	s.countdownTo = time.Time{}
	s.countdownTimer.Stop()
}

func (s *Server) resetTimer() {
	s.stopTimer()
	if s.countdownDuration == 0 {
		return
	}
	s.countdownTo = time.Now().Add(s.countdownDuration)
	s.countdownTimer.Reset(s.countdownDuration)
}
