package main

import (
	"log"
)

type Agent struct {
	sheets *SheetsClient
	gemini *GeminiClient
	wa     *WAManager
}

func NewAgent(sheets *SheetsClient, gemini *GeminiClient, wa *WAManager) *Agent {
	return &Agent{sheets: sheets, gemini: gemini, wa: wa}
}

// HandleIncomingMessage est appelé par whatsapp.go à chaque message reçu.
func (a *Agent) HandleIncomingMessage(from, text string) {
	status, err := a.sheets.GetStatus()
	if err != nil {
		log.Println("Erreur recuperation status:", err)
		return
	}

	// Mono-client: on ignore tout numero different de celui configure
	if status.NumeroClient != "" && from != status.NumeroClient {
		log.Printf("Message ignore (numero %s different du client configure %s)\n", from, status.NumeroClient)
		return
	}

	if !status.BotOperationnel {
		reason := "desactive manuellement"
		if status.AbonnementExpire {
			reason = "abonnement expire"
		} else if status.LimiteAtteinte {
			reason = "limite de tokens mensuelle atteinte"
		}
		log.Printf("Bot non operationnel (%s), message de %s ignore\n", reason, from)
		return
	}

	if err := a.sheets.SaveMessage(from, "user", text); err != nil {
		log.Println("Erreur sauvegarde message user:", err)
	}

	a.wa.SendTyping(from)
	defer a.wa.StopTyping(from)

	history, err := a.sheets.GetHistory(from, 20)
	if err != nil {
		log.Println("Erreur recuperation historique:", err)
	}

	recentBot, err := a.sheets.GetRecentBotMessages(from, 10)
	if err != nil {
		log.Println("Erreur recuperation messages recents:", err)
	}
	var recentTexts []string
	for _, m := range recentBot {
		recentTexts = append(recentTexts, m.Message)
	}

	reply, inputTokens, outputTokens, err := a.gemini.GenerateReply(status.SystemPrompt, history, text, recentTexts)
	if err != nil {
		log.Println("Erreur Gemini:", err)
		return
	}

	// Anti-repetition: si la reponse est identique a la derniere envoyee, on regenere une fois
	if len(recentBot) > 0 && isTooSimilar(reply, recentBot[0].Message) {
		log.Println("Reponse trop similaire a la precedente, regeneration...")
		reply2, in2, out2, err2 := a.gemini.GenerateReply(status.SystemPrompt, history, text, recentTexts)
		if err2 == nil {
			reply = reply2
			inputTokens += in2
			outputTokens += out2
		}
	}

	newStatus, err := a.sheets.UpdateTokens(inputTokens, outputTokens)
	if err != nil {
		log.Println("Erreur mise a jour tokens:", err)
	}

	if err := a.wa.SendText(from, reply); err != nil {
		log.Println("Erreur envoi message WhatsApp:", err)
		return
	}

	if err := a.sheets.SaveMessage(from, "bot", reply); err != nil {
		log.Println("Erreur sauvegarde message bot:", err)
	}

	if newStatus != nil && newStatus.LimiteAtteinte {
		log.Println("⚠️ Limite de tokens mensuelle atteinte, le bot va se desactiver automatiquement pour les prochains messages")
	}
}
