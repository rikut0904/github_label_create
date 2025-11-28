package usecase

import (
	"context"
	"log"

	"github-setup-app/domain/entity"
	"github-setup-app/domain/repository"
)

type SetupRepositoryUseCase struct {
	githubRepo       repository.GitHubRepository
	appID            string
	appPrivateKey    string
}

func NewSetupRepositoryUseCase(githubRepo repository.GitHubRepository, appID, appPrivateKey string) *SetupRepositoryUseCase {
	return &SetupRepositoryUseCase{
		githubRepo:    githubRepo,
		appID:         appID,
		appPrivateKey: appPrivateKey,
	}
}

func (uc *SetupRepositoryUseCase) Execute(ctx context.Context, repo entity.Repository) error {
	log.Printf("Setting up repository: %s/%s", repo.Owner, repo.Name)

	// シークレットを登録
	if err := uc.createSecrets(ctx, repo); err != nil {
		log.Printf("Error creating secrets: %v", err)
		return err
	}

	// テンプレートファイルを一括作成
	if err := uc.createTemplateFiles(ctx, repo); err != nil {
		log.Printf("Error creating template files: %v", err)
		return err
	}

	log.Printf("Repository setup completed: %s/%s", repo.Owner, repo.Name)
	return nil
}

func (uc *SetupRepositoryUseCase) createSecrets(ctx context.Context, repo entity.Repository) error {
	log.Printf("Creating secrets for repository: %s/%s", repo.Owner, repo.Name)

	// APP_ID を登録
	if err := uc.githubRepo.CreateSecret(ctx, repo, "APP_ID", uc.appID); err != nil {
		return err
	}
	log.Printf("Created APP_ID secret")

	// APP_PRIVATE_KEY を登録
	if err := uc.githubRepo.CreateSecret(ctx, repo, "APP_PRIVATE_KEY", uc.appPrivateKey); err != nil {
		return err
	}
	log.Printf("Created APP_PRIVATE_KEY secret")

	return nil
}

func (uc *SetupRepositoryUseCase) createTemplateFiles(ctx context.Context, repo entity.Repository) error {
	log.Printf("Creating template files for repository: %s/%s", repo.Owner, repo.Name)

	// ワークフローファイルを最後にpushするため、順番を調整
	files := []entity.FileContent{
		entity.DefaultLicenseFile(),
		entity.DefaultContributingFile(),
		entity.DefaultSetupLabelsWorkflow(), // 最後
	}

	// 各ファイルを個別に作成
	if err := uc.githubRepo.CreateFiles(ctx, repo, files, "Add Template"); err != nil {
		return err
	}

	log.Printf("Created all template files")
	return nil
}

func (uc *SetupRepositoryUseCase) DeleteWorkflow(ctx context.Context, repo entity.Repository) error {
	log.Printf("Deleting workflow file: %s/%s", repo.Owner, repo.Name)

	workflowPath := ".github/workflows/setup-labels.yml"
	if err := uc.githubRepo.DeleteWorkflowFile(ctx, repo, workflowPath); err != nil {
		return err
	}

	log.Printf("Workflow file deleted successfully: %s/%s", repo.Owner, repo.Name)
	return nil
}
