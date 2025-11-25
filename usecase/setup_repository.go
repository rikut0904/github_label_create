package usecase

import (
	"context"
	"log"

	"github-setup-app/domain/entity"
	"github-setup-app/domain/repository"
)

type SetupRepositoryUseCase struct {
	githubRepo repository.GitHubRepository
}

func NewSetupRepositoryUseCase(githubRepo repository.GitHubRepository) *SetupRepositoryUseCase {
	return &SetupRepositoryUseCase{
		githubRepo: githubRepo,
	}
}

func (uc *SetupRepositoryUseCase) Execute(ctx context.Context, repo entity.Repository) error {
	log.Printf("Setting up repository: %s/%s", repo.Owner, repo.Name)

	// ワークフローファイルを作成
	if err := uc.createWorkflow(ctx, repo); err != nil {
		log.Printf("Error creating workflow: %v", err)
	}

	log.Printf("Repository setup completed: %s/%s", repo.Owner, repo.Name)
	return nil
}

func (uc *SetupRepositoryUseCase) createWorkflow(ctx context.Context, repo entity.Repository) error {
	workflow := entity.DefaultSetupLabelsWorkflow()

	if err := uc.githubRepo.CreateFile(ctx, repo, workflow); err != nil {
		return err
	}

	log.Printf("Created workflow file")
	return nil
}

func (uc *SetupRepositoryUseCase) DeleteRepository(ctx context.Context, repo entity.Repository) error {
	log.Printf("Deleting repository: %s/%s", repo.Owner, repo.Name)

	if err := uc.githubRepo.DeleteRepository(ctx, repo); err != nil {
		return err
	}

	log.Printf("Repository deleted successfully: %s/%s", repo.Owner, repo.Name)
	return nil
}
