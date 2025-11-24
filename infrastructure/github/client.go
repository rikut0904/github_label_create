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

func (c *GitHubClient) ListLabels(ctx context.Context, repo entity.Repository) ([]entity.Label, error) {
	client, err := c.getClient(repo.InstallationID)
	if err != nil {
		return nil, err
	}

	ghLabels, _, err := client.Issues.ListLabels(ctx, repo.Owner, repo.Name, nil)
	if err != nil {
		return nil, err
	}

	labels := make([]entity.Label, len(ghLabels))
	for i, l := range ghLabels {
		labels[i] = entity.Label{
			Name:        l.GetName(),
			Color:       l.GetColor(),
			Description: l.GetDescription(),
		}
	}

	return labels, nil
}

func (c *GitHubClient) DeleteLabel(ctx context.Context, repo entity.Repository, name string) error {
	client, err := c.getClient(repo.InstallationID)
	if err != nil {
		return err
	}

	_, err = client.Issues.DeleteLabel(ctx, repo.Owner, repo.Name, name)
	return err
}

func (c *GitHubClient) CreateLabel(ctx context.Context, repo entity.Repository, label entity.Label) error {
	client, err := c.getClient(repo.InstallationID)
	if err != nil {
		return err
	}

	_, _, err = client.Issues.CreateLabel(ctx, repo.Owner, repo.Name, &github.Label{
		Name:        github.String(label.Name),
		Color:       github.String(label.Color),
		Description: github.String(label.Description),
	})

	return err
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
