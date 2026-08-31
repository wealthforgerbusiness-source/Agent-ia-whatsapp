package main

import (
	"context"
	"encoding/base64"
	"log"
	"sync"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite"
)

type WAManager struct {
	client *whatsmeow.Client
	mu     sync.RWMutex
	qrCode string
	status string

	onMessage func(from, text string)
}

func NewWAManager() (*WAManager, error) {
	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New("sqlite", "file:whatsapp_session.db?_foreign_keys=on", dbLog)
	if err != nil {
		return nil, err
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		return nil, err
	}

	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	return &WAManager{
		client: client,
		status: "disconnected",
	}, nil
}

func (w *WAManager) SetMessageHandler(handler func(from, text string)) {
	w.onMessage = handler
}

func (w *WAManager) setStatus(s string) {
	w.mu.Lock()
	w.status = s
	w.mu.Unlock()
}

func (w *WAManager) GetStatus() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.status
}

func (w *WAManager) setQR(code string) {
	png, err := qrcode.Encode(code, qrcode.Medium, 300)
	if err != nil {
		log.Println("Erreur generation QR:", err)
		return
	}
	w.mu.Lock()
	w.qrCode = base64.StdEncoding.EncodeToString(png)
	w.mu.Unlock()
}

func (w *WAManager) GetQRBase64() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.qrCode
}

func (w *WAManager) clearQR() {
	w.mu.Lock()
	w.qrCode = ""
	w.mu.Unlock()
}

func (w *WAManager) registerHandlers() {
	w.client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.IsFromMe {
				return
			}
			text := v.Message.GetConversation()
			if text == "" && v.Message.GetExtendedTextMessage() != nil {
				text = v.Message.GetExtendedTextMessage().GetText()
			}
			if text == "" {
				return
			}
			from := v.Info.Sender.User
			if w.onMessage != nil {
				w.onMessage(from, text)
			}

		case *events.Connected:
			w.setStatus("connecte")
			w.clearQR()
			log.Println("WhatsApp connecte")

		case *events.LoggedOut:
			w.setStatus("deconnecte")
			log.Println("Session WhatsApp expiree ou deconnectee, generation d'un nouveau QR")
			go func() {
				if err := w.Start(); err != nil {
					log.Println("Erreur relance apres logout:", err)
				}
			}()

		case *events.Disconnected:
			w.setStatus("deconnecte")
			log.Println("WhatsApp deconnecte, whatsmeow va tenter une reconnexion automatique")
		}
	})
}

// Start lance la connexion. Si aucune session existante ou si elle a expire,
// genere un QR code que l'interface admin peut afficher (polling sur GetQRBase64).
func (w *WAManager) Start() error {
	w.registerHandlers()

	if w.client.Store.ID == nil {
		w.setStatus("attente_scan_qr")
		qrChan, _ := w.client.GetQRChannel(context.Background())
		err := w.client.Connect()
		if err != nil {
			return err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				w.setQR(evt.Code)
				log.Println("Nouveau QR code genere, a scanner depuis l'interface admin")
			} else {
				log.Println("Evenement QR:", evt.Event)
			}
		}
	} else {
		err := w.client.Connect()
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *WAManager) jid(numero string) types.JID {
	return types.NewJID(numero, types.DefaultUserServer)
}

func (w *WAManager) SendTyping(numero string) {
	_ = w.client.SendChatPresence(w.jid(numero), types.ChatPresenceComposing, types.ChatPresenceMediaText)
}

func (w *WAManager) StopTyping(numero string) {
	_ = w.client.SendChatPresence(w.jid(numero), types.ChatPresencePaused, types.ChatPresenceMediaText)
}

func (w *WAManager) SendText(numero, text string) error {
	_, err := w.client.SendMessage(context.Background(), w.jid(numero), &waE2E.Message{
		Conversation: proto.String(text),
	})
	return err
}
