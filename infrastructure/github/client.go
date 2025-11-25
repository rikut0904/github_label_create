package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v57/github"

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
