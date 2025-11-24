package repository

import (
	"context"

	"github-setup-app/domain/entity"
)

type GitHubRepository interface {
	ListLabels(ctx context.Context, repo entity.Repository) ([]entity.Label, error)
	DeleteLabel(ctx context.Context, repo entity.Repository, name string) error
	CreateLabel(ctx context.Context, repo entity.Repository, label entity.Label) error
	CreateFile(ctx context.Context, repo entity.Repository, workflow entity.Workflow) error
}
