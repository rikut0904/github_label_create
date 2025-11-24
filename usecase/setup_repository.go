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

	// 既存のラベルを削除
	if err := uc.deleteExistingLabels(ctx, repo); err != nil {
		log.Printf("Error deleting existing labels: %v", err)
	}

	// 新しいラベルを作成
	if err := uc.createLabels(ctx, repo); err != nil {
		log.Printf("Error creating labels: %v", err)
	}

	// ワークフローファイルを作成
	if err := uc.createWorkflow(ctx, repo); err != nil {
		log.Printf("Error creating workflow: %v", err)
	}

	log.Printf("Repository setup completed: %s/%s", repo.Owner, repo.Name)
	return nil
}

func (uc *SetupRepositoryUseCase) deleteExistingLabels(ctx context.Context, repo entity.Repository) error {
	labels, err := uc.githubRepo.ListLabels(ctx, repo)
	if err != nil {
		return err
	}

	for _, label := range labels {
		if err := uc.githubRepo.DeleteLabel(ctx, repo, label.Name); err != nil {
			log.Printf("Error deleting label %s: %v", label.Name, err)
		}
	}

	return nil
}

func (uc *SetupRepositoryUseCase) createLabels(ctx context.Context, repo entity.Repository) error {
	labels := entity.DefaultLabels()

	for _, label := range labels {
		if err := uc.githubRepo.CreateLabel(ctx, repo, label); err != nil {
			log.Printf("Error creating label %s: %v", label.Name, err)
		} else {
			log.Printf("Created label: %s", label.Name)
		}
	}

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
