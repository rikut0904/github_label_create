package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	"github-setup-app/infrastructure/github"
	"github-setup-app/interface/handler"
	"github-setup-app/usecase"
)

func main() {
	// .env を読み込んでローカル・Docker双方で同じ挙動にする
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: could not load .env file: %v", err)
	}

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

	privateKeyEnv := os.Getenv("GITHUB_PRIVATE_KEY")
	if privateKeyEnv == "" {
		log.Fatal("GITHUB_PRIVATE_KEY is required")
	}
	privateKeyEnv = strings.ReplaceAll(privateKeyEnv, `\n`, "\n")
	privateKey := []byte(privateKeyEnv)
	if !strings.Contains(privateKeyEnv, "BEGIN") {
		decoded, err := base64.StdEncoding.DecodeString(privateKeyEnv)
		if err != nil {
			log.Fatal("GITHUB_PRIVATE_KEY must be PEM or base64 encoded PEM content")
		}
		privateKey = decoded
	}

	webhookSecret := os.Getenv("WEBHOOK_SECRET")

	// Infrastructure
	githubClient := github.NewGitHubClient(appID, privateKey)

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
