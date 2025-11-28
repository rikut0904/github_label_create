# アーキテクチャ

このドキュメントでは、GitHub Setup App のアーキテクチャと設計思想を説明します。

## 概要

GitHub Setup App は**クリーンアーキテクチャ**に基づいて設計されています。

### 設計原則

1. **依存性逆転の原則**: 上位層が下位層に依存しない
2. **単一責任の原則**: 各モジュールは1つの責任のみを持つ
3. **疎結合**: モジュール間の結合度を最小化
4. **テスタビリティ**: インターフェースを通じてモックが容易

---

## レイヤー構造

```
┌─────────────────────────────────────────┐
│          main.go (エントリーポイント)      │
│          - DI (依存性注入)                │
│          - サーバー起動                    │
└─────────────────────────────────────────┘
                    │
                    ↓
┌─────────────────────────────────────────┐
│      Interface Layer (インターフェース層)   │
│      - handler/webhook.go                │
│      - handler/health.go                 │
│      └─ HTTPリクエストの受信・レスポンス    │
└─────────────────────────────────────────┘
                    │
                    ↓
┌─────────────────────────────────────────┐
│      UseCase Layer (ユースケース層)        │
│      - usecase/setup_repository.go       │
│      └─ ビジネスロジック                  │
└─────────────────────────────────────────┘
                    │
                    ↓
┌─────────────────────────────────────────┐
│      Domain Layer (ドメイン層)            │
│      - entity/ (エンティティ)              │
│      - repository/ (インターフェース)       │
│      └─ ビジネスルール                    │
└─────────────────────────────────────────┘
                    ↑
                    │ 実装
┌─────────────────────────────────────────┐
│   Infrastructure Layer (インフラ層)        │
│   - infrastructure/github/client.go      │
│   └─ 外部サービスとの通信                 │
└─────────────────────────────────────────┘
```

---

## 各レイヤーの責務

### 1. Interface Layer (interface/)

**責務**: 外部からの入力を受け取り、ユースケースを呼び出す

**ファイル**:
- `handler/webhook.go`: GitHub Webhook の受信
- `handler/health.go`: ヘルスチェック

**特徴**:
- HTTPリクエスト/レスポンスの処理
- Webhook署名の検証
- イベントタイプごとの処理分岐
- ユースケースを非同期で実行（goroutine）

```go
// 例: Webhook ハンドラー
func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. リクエストを受け取る
    // 2. 署名を検証
    // 3. イベントをパース
    // 4. ユースケースを呼び出す
    go func() {
        h.setupUseCase.Execute(ctx, repo)
    }()
}
```

### 2. UseCase Layer (usecase/)

**責務**: ビジネスロジックの実行

**ファイル**:
- `setup_repository.go`: リポジトリセットアップのオーケストレーション

**特徴**:
- ドメインオブジェクトを操作
- リポジトリインターフェースを通じて永続化
- トランザクション境界

```go
// 例: セットアップユースケース
func (uc *SetupRepositoryUseCase) Execute(ctx context.Context, repo entity.Repository) error {
    // 1. シークレット登録
    if err := uc.createSecrets(ctx, repo); err != nil {
        return err
    }

    // 2. テンプレートファイル作成（LICENSE、CONTRIBUTING.md、ワークフロー）
    if err := uc.createTemplateFiles(ctx, repo); err != nil {
        return err
    }

    return nil
}
```

### 3. Domain Layer (domain/)

**責務**: ビジネスルールの定義

**ファイル**:
- `entity/`: データ構造の定義
  - `repository.go`: Repository エンティティ
  - `workflow.go`: Workflow エンティティとテンプレートファイル（LICENSE、CONTRIBUTING.md）
  - `label.go`: Label エンティティ
- `repository/`: インターフェースの定義
  - `github_repository.go`: GitHubリポジトリインターフェース

**特徴**:
- 他のレイヤーに依存しない
- ビジネスルールをカプセル化
- インターフェースを定義（実装は持たない）

```go
// 例: リポジトリインターフェース
type GitHubRepository interface {
    CreateFile(ctx context.Context, repo entity.Repository, file entity.FileContent) error
    CreateFiles(ctx context.Context, repo entity.Repository, files []entity.FileContent, commitMessage string) error
    DeleteWorkflowFile(ctx context.Context, repo entity.Repository, path string) error
    CreateSecret(ctx context.Context, repo entity.Repository, secretName, secretValue string) error
}
```

### 4. Infrastructure Layer (infrastructure/)

**責務**: 外部サービスとの通信

**ファイル**:
- `github/client.go`: GitHub API クライアント

**特徴**:
- ドメイン層のインターフェースを実装
- 外部APIとの通信を隠蔽
- エラーハンドリング

```go
// 例: GitHub クライアント
type GitHubClient struct {
    appID      int64
    privateKey []byte
}

// 単一ファイルを作成（高レベルAPI）
func (c *GitHubClient) CreateFile(ctx context.Context, repo entity.Repository, file entity.FileContent) error {
    client, err := c.getClient(repo.InstallationID)
    // GitHub Repositories.CreateFile API を呼び出す
}

// 複数ファイルを作成（高レベルAPIを複数回実行）
func (c *GitHubClient) CreateFiles(ctx context.Context, repo entity.Repository, files []entity.FileContent, commitMessage string) error {
    // 各ファイルを個別に作成（空のリポジトリでも動作）
    for _, file := range files {
        if err := c.CreateFile(ctx, repo, file); err != nil {
            return err
        }
    }
    return nil
}
```

---

## データフロー

### 1. リポジトリ作成時

```
GitHub
  ↓ repository.created イベント
Webhook Handler (interface)
  ↓ repo entity
Setup UseCase (usecase)
  ↓ CreateSecret, CreateFile
GitHub Client (infrastructure)
  ↓ API呼び出し
GitHub API
```

### 2. ワークフロー完了時

```
GitHub
  ↓ workflow_run.completed イベント
Webhook Handler (interface)
  ↓ repo entity
Setup UseCase (usecase)
  ↓ DeleteWorkflowFile
GitHub Client (infrastructure)
  ↓ API呼び出し
GitHub API
```

---

## 依存性注入 (DI)

`main.go` で依存性を組み立てます:

```go
func main() {
    // 1. Infrastructure 層を作成
    githubClient := github.NewGitHubClient(appID, privateKey)

    // 2. UseCase 層を作成（Infrastructure を注入）
    setupUseCase := usecase.NewSetupRepositoryUseCase(githubClient, labelAppID, labelPrivateKey)

    // 3. Interface 層を作成（UseCase を注入）
    webhookHandler := handler.NewWebhookHandler(setupUseCase, webhookSecret)

    // 4. ルーターに登録
    http.HandleFunc("/webhook", webhookHandler.Handle)
}
```

**メリット**:
- テスト時にモックを注入できる
- 実装を簡単に差し替えられる
- 依存関係が明確

---

## セキュリティアーキテクチャ

### 2つの GitHub App を使用

```
┌─────────────────────────────────┐
│  メインApp                        │
│  - Contents: Read & Write        │
│  - Secrets: Read & Write         │
│                                  │
│  秘密鍵: Railway のみに保存       │
└─────────────────────────────────┘
            │
            ↓ シークレット登録
┌─────────────────────────────────┐
│  新規リポジトリ                   │
│  - APP_ID (ラベル操作App)         │
│  - APP_PRIVATE_KEY               │
└─────────────────────────────────┘
            │
            ↓ ワークフロー実行
┌─────────────────────────────────┐
│  ラベル操作App                    │
│  - Issues: Read & Write          │
│                                  │
│  秘密鍵: 各リポジトリに配布       │
└─────────────────────────────────┘
```

**セキュリティメリット**:
1. **権限分離**: 各Appは最小限の権限のみ
2. **秘密鍵の分離**: 強力な権限を持つ秘密鍵はRailwayのみ
3. **影響範囲の限定**: 漏洩時の被害を最小化

---

## エラーハンドリング

### 1. Webhook レベル

```go
// 署名検証エラー → 401 Unauthorized
if !h.verifySignature(payload, signature) {
    http.Error(w, "Invalid signature", http.StatusUnauthorized)
    return
}

// パースエラー → 400 Bad Request
if err := json.Unmarshal(payload, &event); err != nil {
    http.Error(w, "Error parsing payload", http.StatusBadRequest)
    return
}
```

### 2. UseCase レベル

```go
// ビジネスロジックエラー → ログに記録
if err := uc.createSecrets(ctx, repo); err != nil {
    log.Printf("Error creating secrets: %v", err)
    return err
}
```

### 3. Infrastructure レベル

```go
// API エラー → エラーをラップして返す
if err != nil {
    return fmt.Errorf("failed to create secret: %w", err)
}
```

**非同期処理**:
- Webhook ハンドラーはすぐに 200 OK を返す
- 実際の処理は goroutine で非同期実行
- エラーはログに記録

---

## スケーラビリティ

### 水平スケーリング

複数インスタンスで並行実行可能:

```
┌─────────┐
│ GitHub  │
└─────────┘
     │
     ↓ Webhook
┌─────────────────────────┐
│  Load Balancer          │
└─────────────────────────┘
     │        │        │
     ↓        ↓        ↓
┌────────┐┌────────┐┌────────┐
│ App #1 ││ App #2 ││ App #3 │
└────────┘└────────┘└────────┘
```

**特徴**:
- ステートレス: サーバー間で状態を共有しない
- 冪等性: 同じイベントを複数回受信しても安全
- 非同期処理: Webhook をすぐに返すため、タイムアウトしない

---

## パフォーマンス最適化

### 1. 非同期処理

Webhook 受信後、すぐに 200 OK を返し、処理は goroutine で実行:

```go
go func() {
    ctx := context.Background()
    if err := h.setupUseCase.Execute(ctx, repo); err != nil {
        log.Printf("Error: %v", err)
    }
}()

w.WriteHeader(http.StatusOK)
```

### 2. コネクションプーリング

GitHub API クライアントは再利用:

```go
// getClient は installation ごとにクライアントを作成
func (c *GitHubClient) getClient(installationID int64) (*github.Client, error) {
    itr, err := ghinstallation.New(...)
    return github.NewClient(&http.Client{Transport: itr}), nil
}
```

---

## テスト戦略

### 1. 単体テスト

各レイヤーを独立してテスト:

```go
// UseCase のテスト
func TestSetupRepositoryUseCase_Execute(t *testing.T) {
    // モックリポジトリを作成
    mockRepo := &MockGitHubRepository{}

    // ユースケースを作成
    uc := NewSetupRepositoryUseCase(mockRepo, "app-id", "private-key")

    // テスト実行
    err := uc.Execute(context.Background(), testRepo)
    assert.NoError(t, err)
}
```

### 2. 統合テスト

実際の GitHub API を使用してテスト（オプション）

### 3. E2E テスト

テストリポジトリを作成して動作確認

---

## 関連ドキュメント

- [セットアップガイド](./setup.md)
- [権限設定](./permissions.md)
- [開発ガイド](./development.md)
