package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type SheetsClient struct {
	baseURL string
	secret  string
	http    *http.Client
}

func NewSheetsClient(cfg *Config) *SheetsClient {
	return &SheetsClient{
		baseURL: cfg.AppsScriptURL,
		secret:  cfg.AppsScriptSecret,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

type ConfigResponse struct {
	OK                  bool   `json:"ok"`
	BotActif            bool   `json:"bot_actif"`
	DateDebutAbonnement string `json:"date_debut_abonnement"`
	TokensInputTotal    int64  `json:"tokens_input_total"`
	TokensOutputTotal   int64  `json:"tokens_output_total"`
}

func (s *SheetsClient) get(action string, extraParams map[string]string) ([]byte, error) {
	q := url.Values{}
	q.Set("secret", s.secret)
	q.Set("action", action)
	for k, v := range extraParams {
		q.Set(k, v)
	}
	resp, err := s.http.Get(s.baseURL + "?" + q.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *SheetsClient) post(payload map[string]interface{}) ([]byte, error) {
	payload["secret"] = s.secret
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	resp, err := s.http.Post(s.baseURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *SheetsClient) GetConfig() (*ConfigResponse, error) {
	raw, err := s.get("config", nil)
	if err != nil {
		return nil, err
	}
	var cfg ConfigResponse
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	if !cfg.OK {
		return nil, fmt.Errorf("erreur config: %s", string(raw))
	}
	return &cfg, nil
}

func (s *SheetsClient) CheckPassword(password string) (bool, error) {
	raw, err := s.get("check_password", map[string]string{"password": password})
	if err != nil {
		return false, err
	}
	var res struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return false, err
	}
	return res.OK, nil
}

func (s *SheetsClient) ToggleBot(actif bool) error {
	_, err := s.post(map[string]interface{}{
		"action": "toggle_bot",
		"actif":  actif,
	})
	return err
}

func (s *SheetsClient) ResetAbonnement() error {
	_, err := s.post(map[string]interface{}{
		"action": "reset_abonnement",
	})
	return err
}

func (s *SheetsClient) SaveMessage(numero, role, message string) error {
	_, err := s.post(map[string]interface{}{
		"action":  "save_message",
		"numero":  numero,
		"role":    role,
		"message": message,
	})
	return err
}

func (s *SheetsClient) UpdateTokens(input, output int) (*ConfigResponse, error) {
	raw, err := s.post(map[string]interface{}{
		"action": "update_tokens",
		"input":  input,
		"output": output,
	})
	if err != nil {
		return nil, err
	}
	var res ConfigResponse
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type RecentMessage struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

func (s *SheetsClient) GetRecentBotMessages(numero string, limit int) ([]RecentMessage, error) {
	raw, err := s.get("recent_messages", map[string]string{
		"numero": numero,
		"limit":  fmt.Sprintf("%d", limit),
	})
	if err != nil {
		return nil, err
	}
	var res struct {
		OK       bool             `json:"ok"`
		Messages []RecentMessage  `json:"messages"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	return res.Messages, nil
}

type HistoryMessage struct {
	Timestamp string `json:"timestamp"`
	Role      string `json:"role"`
	Message   string `json:"message"`
}

func (s *SheetsClient) GetHistory(numero string, limit int) ([]HistoryMessage, error) {
	raw, err := s.get("history", map[string]string{
		"numero": numero,
		"limit":  fmt.Sprintf("%d", limit),
	})
	if err != nil {
		return nil, err
	}
	var res struct {
		OK       bool             `json:"ok"`
		Messages []HistoryMessage `json:"messages"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	return res.Messages, nil
}
