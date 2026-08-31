package main

import (
	"log"
	"os"
)

type Config struct {
	AppsScriptURL    string
	AppsScriptSecret string
	GeminiAPIKey     string
	Port             string
	AdminCookieName  string
}

func LoadConfig() *Config {
	cfg := &Config{
		AppsScriptURL:    os.Getenv("APPS_SCRIPT_URL"),
		AppsScriptSecret: os.Getenv("APPS_SCRIPT_SECRET"),
		GeminiAPIKey:     os.Getenv("GEMINI_API_KEY"),
		Port:             os.Getenv("PORT"),
		AdminCookieName:  "wa_agent_session",
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	if cfg.AppsScriptURL == "" || cfg.AppsScriptSecret == "" || cfg.GeminiAPIKey == "" {
		log.Fatal("Variables manquantes: APPS_SCRIPT_URL, APPS_SCRIPT_SECRET, GEMINI_API_KEY sont obligatoires")
	}

	return cfg
}
