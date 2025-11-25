package repository

import (
	"context"

	"github-setup-app/domain/entity"
)

type GitHubRepository interface {
	CreateFile(ctx context.Context, repo entity.Repository, workflow entity.Workflow) error
	DeleteRepository(ctx context.Context, repo entity.Repository) error
	CreateSecret(ctx context.Context, repo entity.Repository, secretName, secretValue string) error
}
