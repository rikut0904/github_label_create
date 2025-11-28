package github

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v57/github"
	"golang.org/x/crypto/blake2b"
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

func (c *GitHubClient) CreateFile(ctx context.Context, repo entity.Repository, file entity.FileContent) error {
	client, err := c.getClient(repo.InstallationID)
	if err != nil {
		return err
	}

	_, _, err = client.Repositories.CreateFile(ctx, repo.Owner, repo.Name, file.GetPath(), &github.RepositoryContentFileOptions{
		Message: github.String(file.GetMessage()),
		Content: []byte(file.GetContent()),
	})

	return err
}

func (c *GitHubClient) CreateFiles(ctx context.Context, repo entity.Repository, files []entity.FileContent, commitMessage string) error {
	// 各ファイルを個別に作成（空のリポジトリでも動作する）
	for _, file := range files {
		if err := c.CreateFile(ctx, repo, file); err != nil {
			return fmt.Errorf("failed to create file %s: %w", file.GetPath(), err)
		}
	}
	return nil
}

func (c *GitHubClient) DeleteWorkflowFile(ctx context.Context, repo entity.Repository, path string) error {
	client, err := c.getClient(repo.InstallationID)
	if err != nil {
		return err
	}

	// ファイルの SHA を取得
	fileContent, _, _, err := client.Repositories.GetContents(ctx, repo.Owner, repo.Name, path, nil)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// ファイルを削除
	_, _, err = client.Repositories.DeleteFile(ctx, repo.Owner, repo.Name, path, &github.RepositoryContentFileOptions{
		Message: github.String("Remove workflow file after setup completion"),
		SHA:     fileContent.SHA,
	})

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

	// nonce を生成（BLAKE2b hash of ephemeralPublicKey + publicKey）
	var nonce [24]byte
	nonceHash, err := blake2b.New(24, nil)
	if err != nil {
		return nil, err
	}
	nonceHash.Write(ephemeralPublicKey[:])
	nonceHash.Write(publicKey[:])
	copy(nonce[:], nonceHash.Sum(nil))

	// メッセージを暗号化
	// 結果: ephemeralPublicKey (32 bytes) + encrypted message
	encrypted := box.Seal(ephemeralPublicKey[:], message, &nonce, publicKey, ephemeralPrivateKey)

	return encrypted, nil
}
