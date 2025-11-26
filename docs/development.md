# 開発ガイド

このドキュメントでは、GitHub Setup App の開発方法を説明します。

## 開発環境のセットアップ

### 前提条件

- Go 1.24 以上
- Docker & Docker Compose（オプション）
- ngrok（ローカルでWebhookテストする場合）

### 1. リポジトリをクローン

```bash
git clone https://github.com/yourusername/github-setup-app.git
cd github-setup-app
```

### 2. 依存関係のインストール

```bash
go mod download
```

### 3. 環境変数を設定

`.env` ファイルを作成:

```bash
cp .env.example .env
```

`.env` を編集:

```env
# メインApp（リポジトリ操作用）
GITHUB_APP_ID=123456
GITHUB_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----

# ラベル操作専用App
LABEL_APP_ID=789012
LABEL_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----

# Webhook Secret
WEBHOOK_SECRET=your-webhook-secret

# Port
PORT=8080
```

---

## ローカル開発

### 方法1: Go で直接実行

```bash
# 実行
go run main.go

# 別のターミナルでngrokを起動
ngrok http 8080
```

ngrok の URL を GitHub App の Webhook URL に設定:
```
https://xxxx-xx-xxx-xxx-xx.ngrok-free.app/webhook
```

### 方法2: Docker で実行

```bash
# ビルド
docker build -t github-setup-app .

# 実行
docker run -p 8080:8080 --env-file .env github-setup-app

# または Docker Compose で
docker-compose up
```

### 動作確認

```bash
# ヘルスチェック
curl http://localhost:8080/health
# レスポンス: OK
```

---

## コードの構造

このプロジェクトはクリーンアーキテクチャに基づいています。

```
github-setup-app/
├── main.go                          # エントリーポイント、DI
├── domain/                          # ドメイン層
│   ├── entity/                      # エンティティ
│   │   ├── label.go
│   │   ├── repository.go
│   │   └── workflow.go
│   └── repository/                  # リポジトリインターフェース
│       └── github_repository.go
├── usecase/                         # ユースケース層
│   └── setup_repository.go
├── infrastructure/                  # インフラ層
│   └── github/
│       └── client.go                # GitHub API クライアント
├── interface/                       # インターフェース層
│   └── handler/
│       ├── webhook.go               # Webhook ハンドラー
│       └── health.go                # ヘルスチェック
├── docs/                            # ドキュメント
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
├── .env.example
└── .gitignore
```

### 依存関係の方向

```
main.go
  ↓
interface (handler)
  ↓
usecase
  ↓
domain (repository interface)
  ↑ 実装
infrastructure (github client)
```

---

## 主要なファイル

### main.go

- エントリーポイント
- 環境変数の読み込み
- 依存性注入（DI）
- HTTPサーバーの起動

### interface/handler/webhook.go

- Webhook イベントの受信
- 署名検証
- イベントタイプごとの処理分岐
  - `repository.created` → セットアップ処理
  - `workflow_run.completed` → ワークフローファイル削除

### usecase/setup_repository.go

- リポジトリセットアップのビジネスロジック
- シークレット登録
- ワークフローファイル作成
- ワークフローファイル削除

### infrastructure/github/client.go

- GitHub API との通信
- ファイル作成・削除
- シークレット暗号化・登録

### domain/entity/workflow.go

- ワークフローファイルの内容を定義
- `DefaultSetupLabelsWorkflow()` でデフォルト設定を返す

---

## テスト

### 単体テストの実行

```bash
go test ./...
```

### 手動テスト

1. **ローカルサーバーを起動**
   ```bash
   go run main.go
   ```

2. **ngrok でトンネル作成**
   ```bash
   ngrok http 8080
   ```

3. **GitHub App の Webhook URL を ngrok URL に変更**
   ```
   https://xxxx.ngrok-free.app/webhook
   ```

4. **新しいリポジトリを作成**してテスト

5. **ログを確認**
   ```
   Setting up repository: user/repo
   Creating secrets for repository: user/repo
   Created APP_ID secret
   Created APP_PRIVATE_KEY secret
   Created workflow file
   Repository setup completed: user/repo
   ```

---

## デバッグ

### ログの確認

アプリケーションは標準出力にログを出力します:

```bash
# ローカル
go run main.go

# Docker
docker-compose logs -f

# Railway
Railway ダッシュボード → Deployments → View Logs
```

### よくあるエラー

#### 1. シークレット暗号化エラー

```
Error creating secrets: failed to create secret: 422 Bad request
```

**原因**: 秘密鍵の形式が間違っている

**解決**:
- 改行が `\n` になっているか確認
- `-----BEGIN RSA PRIVATE KEY-----` で始まっているか確認

#### 2. ワークフローが 403 エラー

```
HTTP 403: Resource not accessible by integration
```

**原因**: ラベル操作Appがインストールされていない

**解決**: ラベル操作App → Install App → All repositories

#### 3. Webhook 署名エラー

```
Invalid signature
```

**原因**: WEBHOOK_SECRET が間違っている

**解決**: GitHub App の設定と環境変数の WEBHOOK_SECRET を一致させる

---

## ビルド

### ローカルビルド

```bash
go build -o github-setup-app
./github-setup-app
```

### Docker ビルド

```bash
docker build -t github-setup-app .
docker run -p 8080:8080 --env-file .env github-setup-app
```

---

## デプロイ

### Railway

1. GitHub リポジトリを Railway に接続
2. 環境変数を設定
3. 自動的にデプロイされる

### その他のプラットフォーム

- Heroku
- Google Cloud Run
- AWS ECS
- 任意のDockerホスティング

**必要な設定**:
- ポート: `8080`（環境変数 `PORT` で変更可能）
- 環境変数: `GITHUB_APP_ID`, `GITHUB_PRIVATE_KEY`, `LABEL_APP_ID`, `LABEL_PRIVATE_KEY`, `WEBHOOK_SECRET`

---

## コーディング規約

### Go のスタイル

- [Effective Go](https://golang.org/doc/effective_go) に従う
- `gofmt` でフォーマット
- `golint` でリント

### コミットメッセージ

```
<type>: <subject>

<body>

🤖 Generated with Claude Code
Co-Authored-By: Claude <noreply@anthropic.com>
```

**Type**:
- `feat`: 新機能
- `fix`: バグ修正
- `docs`: ドキュメント
- `refactor`: リファクタリング
- `test`: テスト

---

## 関連ドキュメント

- [セットアップガイド](./setup.md)
- [権限設定](./permissions.md)
- [アーキテクチャ](./architecture.md)
