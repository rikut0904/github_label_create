package main

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
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

	// このGitHub App（リポジトリ操作用）
	appIDStr := os.Getenv("GITHUB_APP_ID")
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid GITHUB_APP_ID: %v", err)
	}

	privateKeyEnv := os.Getenv("GITHUB_PRIVATE_KEY")
	if privateKeyEnv == "" {
		log.Fatal("GITHUB_PRIVATE_KEY is required")
	}
	privateKeyEnv, err = normalizePrivateKeyEnv("GITHUB_PRIVATE_KEY", privateKeyEnv)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded GITHUB_PRIVATE_KEY as %s", summarizePrivateKey(privateKeyEnv))
	privateKey := []byte(privateKeyEnv)

	// ラベル操作専用GitHub App
	labelAppIDStr := os.Getenv("LABEL_APP_ID")
	if labelAppIDStr == "" {
		log.Fatal("LABEL_APP_ID is required")
	}

	labelPrivateKeyEnv := os.Getenv("LABEL_PRIVATE_KEY")
	if labelPrivateKeyEnv == "" {
		log.Fatal("LABEL_PRIVATE_KEY is required")
	}
	labelPrivateKeyEnv, err = normalizePrivateKeyEnv("LABEL_PRIVATE_KEY", labelPrivateKeyEnv)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded LABEL_PRIVATE_KEY as %s", summarizePrivateKey(labelPrivateKeyEnv))

	webhookSecret := os.Getenv("WEBHOOK_SECRET")

	// Infrastructure
	githubClient := github.NewGitHubClient(appID, privateKey)

	// UseCase (シークレット登録のため labelAppIDStr と labelPrivateKeyEnv を渡す)
	setupUseCase := usecase.NewSetupRepositoryUseCase(githubClient, labelAppIDStr, labelPrivateKeyEnv)

	// Handler
	webhookHandler := handler.NewWebhookHandler(setupUseCase, webhookSecret)
	healthHandler := handler.NewHealthHandler()

	// Router
	http.HandleFunc("/webhook", webhookHandler.Handle)
	http.HandleFunc("/health", healthHandler.Handle)

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func normalizePrivateKeyEnv(envName, raw string) (string, error) {
	text := sanitizeKeyText(raw)
	if !strings.Contains(text, "BEGIN") || !strings.Contains(text, "PRIVATE KEY") {
		decoded, err := decodeKeyBase64(text)
		if err != nil {
			return "", fmt.Errorf("%s must be PEM or base64 encoded PEM content: %w", envName, err)
		}
		text = sanitizeKeyText(string(decoded))
	}

	if err := validatePEMPrivateKey(text); err != nil {
		return "", fmt.Errorf("%s is not a valid PEM encoded PKCS1 or PKCS8 private key: %w", envName, err)
	}

	return text, nil
}

func sanitizeKeyText(raw string) string {
	text := strings.TrimSpace(raw)
	text = unquoteKeyText(text)
	text = strings.Trim(text, `"'`)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, `\\r\\n`, "\n")
	text = strings.ReplaceAll(text, `\\n`, "\n")
	text = strings.ReplaceAll(text, `\r\n`, "\n")
	text = strings.ReplaceAll(text, `\n`, "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return text
}

func unquoteKeyText(text string) string {
	for {
		unquoted, err := strconv.Unquote(text)
		if err != nil {
			return text
		}
		text = strings.TrimSpace(unquoted)
	}
}

func decodeKeyBase64(text string) ([]byte, error) {
	decoders := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}

	var lastErr error
	for _, enc := range decoders {
		decoded, err := enc.DecodeString(text)
		if err == nil {
			return decoded, nil
		}
		lastErr = err
	}

	return nil, lastErr
}

func validatePEMPrivateKey(text string) error {
	block, _ := pem.Decode([]byte(text))
	if block == nil {
		return fmt.Errorf("PEM block not found")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		if _, err := x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
			return err
		}
		return nil
	case "PRIVATE KEY":
		_, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		return err
	default:
		return fmt.Errorf("unexpected PEM block type %q", block.Type)
	}
}

func summarizePrivateKey(text string) string {
	block, _ := pem.Decode([]byte(text))
	if block == nil {
		return fmt.Sprintf("invalid PEM (%d bytes)", len(text))
	}
	return fmt.Sprintf("%s (%d bytes)", block.Type, len(text))
}
