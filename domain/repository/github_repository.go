package repository

import (
	"context"

	"github-setup-app/domain/entity"
)

type GitHubRepository interface {
	CreateFile(ctx context.Context, repo entity.Repository, workflow entity.Workflow) error
}
