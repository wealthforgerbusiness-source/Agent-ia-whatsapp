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

// Status reflète exactement ce que computeStatus_() renvoie côté Apps Script
type Status struct {
	OK                    bool   `json:"ok"`
	BotActif              bool   `json:"bot_actif"`
	BotOperationnel       bool   `json:"bot_operationnel"`
	NumeroClient          string `json:"numero_client"`
	DateDebutAbonnement   string `json:"date_debut_abonnement"`
	DureeAbonnementJours  int    `json:"duree_abonnement_jours"`
	JoursRestants         int    `json:"jours_restants"`
	AbonnementExpire      bool   `json:"abonnement_expire"`
	TokensInputTotal      int64  `json:"tokens_input_total"`
	TokensOutputTotal     int64  `json:"tokens_output_total"`
	TokensTotal           int64  `json:"tokens_total"`
	TokensLimiteMensuelle int64  `json:"tokens_limite_mensuelle"`
	TokensRestants        int64  `json:"tokens_restants"`
	LimiteAtteinte        bool   `json:"limite_atteinte"`
	SystemPrompt          string `json:"system_prompt"`
	Blocked               bool   `json:"blocked"`
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

func parseStatus_(raw []byte) (*Status, error) {
	var status Status
	if err := json.Unmarshal(raw, &status); err != nil {
		return nil, fmt.Errorf("erreur parsing status: %s", string(raw))
	}
	if !status.OK {
		return nil, fmt.Errorf("erreur reponse sheets: %s", string(raw))
	}
	return &status, nil
}

func (s *SheetsClient) GetStatus() (*Status, error) {
	raw, err := s.get("config", nil)
	if err != nil {
		return nil, err
	}
	return parseStatus_(raw)
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

func (s *SheetsClient) ToggleBot(actif bool) (*Status, error) {
	raw, err := s.post(map[string]interface{}{
		"action": "toggle_bot",
		"actif":  actif,
	})
	if err != nil {
		return nil, err
	}
	return parseStatus_(raw)
}

func (s *SheetsClient) ResetAbonnement() (*Status, error) {
	raw, err := s.post(map[string]interface{}{
		"action": "reset_abonnement",
	})
	if err != nil {
		return nil, err
	}
	return parseStatus_(raw)
}

func (s *SheetsClient) SetNumeroClient(numero string) error {
	_, err := s.post(map[string]interface{}{
		"action": "set_numero_client",
		"numero": numero,
	})
	return err
}

func (s *SheetsClient) SetSystemPrompt(prompt string) error {
	_, err := s.post(map[string]interface{}{
		"action": "set_system_prompt",
		"prompt": prompt,
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

func (s *SheetsClient) UpdateTokens(input, output int) (*Status, error) {
	raw, err := s.post(map[string]interface{}{
		"action": "update_tokens",
		"input":  input,
		"output": output,
	})
	if err != nil {
		return nil, err
	}
	return parseStatus_(raw)
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
		OK       bool            `json:"ok"`
		Messages []RecentMessage `json:"messages"`
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
