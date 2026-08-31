package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	sessions   = map[string]time.Time{}
	sessionsMu sync.Mutex
)

func generateToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func createSession() string {
	token := generateToken()
	sessionsMu.Lock()
	sessions[token] = time.Now().Add(24 * time.Hour)
	sessionsMu.Unlock()
	return token
}

func isValidSession(token string) bool {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	expiry, ok := sessions[token]
	if !ok {
		return false
	}
	if time.Now().After(expiry) {
		delete(sessions, token)
		return false
	}
	return true
}

func main() {
	cfg := LoadConfig()
	sheets := NewSheetsClient(cfg)
	gemini := NewGeminiClient(cfg)

	wa, err := NewWAManager()
	if err != nil {
		log.Fatal("Erreur init WhatsApp manager:", err)
	}

	agent := NewAgent(sheets, gemini, wa)
	wa.SetMessageHandler(agent.HandleIncomingMessage)

	go func() {
		if err := wa.Start(); err != nil {
			log.Println("Erreur demarrage WhatsApp:", err)
		}
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		handleLogin(w, r, sheets, cfg)
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cfg.AdminCookieName)
		if err == nil {
			sessionsMu.Lock()
			delete(sessions, cookie.Value)
			sessionsMu.Unlock()
		}
		http.SetCookie(w, &http.Cookie{Name: cfg.AdminCookieName, Value: "", Path: "/", MaxAge: -1})
		http.Redirect(w, r, "/login", http.StatusFound)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(r, cfg) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(dashboardHTML))
	})

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(r, cfg) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		status, err := sheets.GetStatus()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := map[string]interface{}{
			"bot_actif":               status.BotActif,
			"bot_operationnel":        status.BotOperationnel,
			"numero_client":           status.NumeroClient,
			"jours_restants":          status.JoursRestants,
			"abonnement_expire":       status.AbonnementExpire,
			"tokens_total":            status.TokensTotal,
			"tokens_limite_mensuelle": status.TokensLimiteMensuelle,
			"tokens_restants":         status.TokensRestants,
			"limite_atteinte":         status.LimiteAtteinte,
			"whatsapp_status":         wa.GetStatus(),
			"qr_code":                 wa.GetQRBase64(),
		}
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api/toggle", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(r, cfg) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var body struct {
			Actif bool `json:"actif"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		status, err := sheets.ToggleBot(body.Actif)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(status)
	})

	mux.HandleFunc("/api/reset", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(r, cfg) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		status, err := sheets.ResetAbonnement()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(status)
	})

	log.Println("Serveur demarre sur le port " + cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}

func requireAuth(r *http.Request, cfg *Config) bool {
	cookie, err := r.Cookie(cfg.AdminCookieName)
	if err != nil {
		return false
	}
	return isValidSession(cookie.Value)
}

func handleLogin(w http.ResponseWriter, r *http.Request, sheets *SheetsClient, cfg *Config) {
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(loginHTML))
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		password := r.FormValue("password")

		ok, err := sheets.CheckPassword(password)
		if err != nil || !ok {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(loginHTML + `<script>document.querySelector('.error').style.display='block';</script>`))
			return
		}

		token := createSession()
		http.SetCookie(w, &http.Cookie{
			Name:     cfg.AdminCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   86400,
		})
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
