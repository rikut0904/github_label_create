package github

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v57/github"
	"golang.org/x/crypto/nacl/box"

	"github-setup-app/domain/entity"
)

type GitHubClient struct {
	appID      int64
	privateKey []byte
}

func NewGitHubClient(appID int64, privateKey []byte) *GitHubClient {
	return &GitHubClient{
		appID:      appID,
		privateKey: privateKey,
	}
}

func (c *GitHubClient) getClient(installationID int64) (*github.Client, error) {
	itr, err := ghinstallation.New(
		http.DefaultTransport,
		c.appID,
		installationID,
		c.privateKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create installation transport: %w", err)
	}

	return github.NewClient(&http.Client{Transport: itr}), nil
}

func (c *GitHubClient) CreateFile(ctx context.Context, repo entity.Repository, workflow entity.Workflow) error {
	client, err := c.getClient(repo.InstallationID)
	if err != nil {
		return err
	}

	_, _, err = client.Repositories.CreateFile(ctx, repo.Owner, repo.Name, workflow.Path, &github.RepositoryContentFileOptions{
		Message: github.String(workflow.Message),
		Content: []byte(workflow.Content),
	})

	return err
}

func (c *GitHubClient) DeleteRepository(ctx context.Context, repo entity.Repository) error {
	client, err := c.getClient(repo.InstallationID)
	if err != nil {
		return err
	}

	_, err = client.Repositories.Delete(ctx, repo.Owner, repo.Name)
	return err
}

func (c *GitHubClient) CreateSecret(ctx context.Context, repo entity.Repository, secretName, secretValue string) error {
	client, err := c.getClient(repo.InstallationID)
	if err != nil {
		return err
	}

	// リポジトリの公開鍵を取得
	publicKey, _, err := client.Actions.GetRepoPublicKey(ctx, repo.Owner, repo.Name)
	if err != nil {
		return fmt.Errorf("failed to get repo public key: %w", err)
	}

	// シークレットを暗号化
	encryptedSecret, err := encryptSecret(publicKey.GetKey(), secretValue)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// シークレットを作成/更新
	encryptedSecretPayload := &github.EncryptedSecret{
		Name:           secretName,
		KeyID:          publicKey.GetKeyID(),
		EncryptedValue: encryptedSecret,
	}

	_, err = client.Actions.CreateOrUpdateRepoSecret(ctx, repo.Owner, repo.Name, encryptedSecretPayload)
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// encryptSecret は libsodium sealed box を使ってシークレットを暗号化
func encryptSecret(publicKeyStr, secret string) (string, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %w", err)
	}

	var publicKey [32]byte
	copy(publicKey[:], publicKeyBytes)

	secretBytes := []byte(secret)

	// nacl/box を使って暗号化（libsodium sealed box と互換）
	encrypted, err := sealBox(secretBytes, &publicKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// sealBox は libsodium の crypto_box_seal と互換性のある暗号化を実行
func sealBox(message []byte, publicKey *[32]byte) ([]byte, error) {
	// 一時的な鍵ペアを生成
	ephemeralPublicKey, ephemeralPrivateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// nonce を生成（ephemeralPublicKey + publicKey のハッシュ）
	var nonce [24]byte
	copy(nonce[:], ephemeralPublicKey[:24])

	// メッセージを暗号化
	encrypted := box.Seal(ephemeralPublicKey[:], message, &nonce, publicKey, ephemeralPrivateKey)

	return encrypted, nil
}
