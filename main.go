package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github-setup-app/infrastructure/github"
	"github-setup-app/interface/handler"
	"github-setup-app/usecase"
)

func main() {
	// 環境変数の読み込み
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appIDStr := os.Getenv("GITHUB_APP_ID")
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid GITHUB_APP_ID: %v", err)
	}

	privateKey := os.Getenv("GITHUB_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("GITHUB_PRIVATE_KEY is required")
	}

	webhookSecret := os.Getenv("WEBHOOK_SECRET")

	// Infrastructure
	githubClient := github.NewGitHubClient(appID, []byte(privateKey))

	// UseCase
	setupUseCase := usecase.NewSetupRepositoryUseCase(githubClient)

	// Handler
	webhookHandler := handler.NewWebhookHandler(setupUseCase, webhookSecret)
	healthHandler := handler.NewHealthHandler()

	// Router
	http.HandleFunc("/webhook", webhookHandler.Handle)
	http.HandleFunc("/health", healthHandler.Handle)

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
