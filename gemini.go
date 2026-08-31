package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const GeminiModel = "gemini-2.0-flash"

type GeminiClient struct {
	apiKey string
	http   *http.Client
}

func NewGeminiClient(cfg *Config) *GeminiClient {
	return &GeminiClient{
		apiKey: cfg.GeminiAPIKey,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiRequest struct {
	Contents          []geminiContent `json:"contents"`
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
}

type geminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	UsageMetadata geminiUsage `json:"usageMetadata"`
}

// GenerateReply appelle Gemini avec le system prompt business (venant du Sheet) + historique + message utilisateur.
// recentBotMessages sert à demander explicitement à l'IA de ne pas répéter une formulation déjà utilisée.
func (g *GeminiClient) GenerateReply(systemPrompt string, history []HistoryMessage, userMessage string, recentBotMessages []string) (string, int, int, error) {
	var contents []geminiContent

	for _, h := range history {
		role := "user"
		if h.Role == "bot" {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: h.Message}},
		})
	}

	contents = append(contents, geminiContent{
		Role:  "user",
		Parts: []geminiPart{{Text: userMessage}},
	})

	systemText := systemPrompt
	if len(recentBotMessages) > 0 {
		systemText += "\n\nIMPORTANT: Ne répète jamais mot pour mot une de tes réponses précédentes. Voici tes derniers messages envoyés à ce client, reformule différemment si le sujet revient:\n- " + strings.Join(recentBotMessages, "\n- ")
	}

	reqBody := geminiRequest{
		Contents: contents,
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemText}},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, 0, err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", GeminiModel, g.apiKey)

	resp, err := g.http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, 0, err
	}

	var gr geminiResponse
	if err := json.Unmarshal(raw, &gr); err != nil {
		return "", 0, 0, fmt.Errorf("erreur parsing reponse gemini: %s", string(raw))
	}

	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return "", 0, 0, fmt.Errorf("reponse gemini vide: %s", string(raw))
	}

	text := gr.Candidates[0].Content.Parts[0].Text
	return text, gr.UsageMetadata.PromptTokenCount, gr.UsageMetadata.CandidatesTokenCount, nil
}

// isTooSimilar compare grossièrement deux messages pour éviter d'envoyer 2x la même réponse exacte.
func isTooSimilar(a, b string) bool {
	a = strings.TrimSpace(strings.ToLower(a))
	b = strings.TrimSpace(strings.ToLower(b))
	return a == b
}
