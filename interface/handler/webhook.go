package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/google/go-github/v57/github"

	"github-setup-app/domain/entity"
	"github-setup-app/usecase"
)

type WebhookHandler struct {
	setupUseCase  *usecase.SetupRepositoryUseCase
	webhookSecret string
}

func NewWebhookHandler(setupUseCase *usecase.SetupRepositoryUseCase, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{
		setupUseCase:  setupUseCase,
		webhookSecret: webhookSecret,
	}
}

func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	if h.webhookSecret != "" {
		signature := r.Header.Get("X-Hub-Signature-256")
		if !h.verifySignature(payload, signature) {
			log.Printf("Invalid signature")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "repository" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var event github.RepositoryEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Printf("Error parsing payload: %v", err)
		http.Error(w, "Error parsing payload", http.StatusBadRequest)
		return
	}

	if event.GetAction() != "created" {
		w.WriteHeader(http.StatusOK)
		return
	}

	repo := entity.Repository{
		Owner:          event.GetRepo().GetOwner().GetLogin(),
		Name:           event.GetRepo().GetName(),
		InstallationID: event.GetInstallation().GetID(),
	}

	go func() {
		if err := h.setupUseCase.Execute(r.Context(), repo); err != nil {
			log.Printf("Error setting up repository: %v", err)
		}
	}()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Processing"))
}

func (h *WebhookHandler) verifySignature(payload []byte, signature string) bool {
	if len(signature) < 7 || signature[:7] != "sha256=" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature[7:]), []byte(expectedMAC))
}
