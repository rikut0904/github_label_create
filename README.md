# GitHub Setup App

リポジトリ作成時にラベルとワークフローを自動設定するGitHub App。

## 機能

- リポジトリ作成時に自動実行
- 既存ラベルを削除し、カスタムラベルを作成
- setup-labels ワークフローファイルを追加
- setup-labels ワークフロー完了後、自動的にリポジトリを削除

## アーキテクチャ

クリーンアーキテクチャで構成されています。

```
github-setup-app/
├── main.go                          # エントリーポイント、DI
├── domain/
│   ├── entity/                      # エンティティ
│   │   ├── label.go
│   │   ├── repository.go
│   │   └── workflow.go
│   └── repository/                  # リポジトリインターフェース
│       └── github_repository.go
├── usecase/                         # ユースケース
│   └── setup_repository.go
├── infrastructure/                  # 外部サービス実装
│   └── github/
│       └── client.go
├── interface/                       # アダプター
│   └── handler/
│       ├── webhook.go
│       └── health.go
├── go.mod
├── Dockerfile
└── .env.example
```

## セットアップ

### 1. GitHub App を2つ作成

#### 1-1. メインApp（リポジトリ操作用）

1. GitHub Settings → Developer settings → GitHub Apps → New GitHub App
2. 以下を設定:
   - **App name**: `Repository Setup App` (任意の名前)
   - **Homepage URL**: Railway のURL（後で設定可）
   - **Webhook URL**: `https://your-app.railway.app/webhook`
   - **Webhook secret**: 任意の文字列（メモしておく）

3. Permissions:
   - **Repository permissions**:
     - Administration: Read and write (リポジトリ削除に必要)
     - Contents: Read and write (ワークフローファイル作成に必要)
     - Metadata: Read-only
     - Secrets: Read and write (シークレット登録に必要)
   - **Subscribe to events**:
     - Repository
     - Workflow run

4. 作成後:
   - App ID をメモ (`GITHUB_APP_ID`)
   - Private key を生成してダウンロード (`GITHUB_PRIVATE_KEY`)

#### 1-2. ラベル操作専用App

1. GitHub Settings → Developer settings → GitHub Apps → New GitHub App
2. 以下を設定:
   - **App name**: `Label Manager` (任意の名前)
   - **Homepage URL**: 任意のURL
   - **Webhook**: **チェックを外す**（このAppはWebhook不要）

3. Permissions:
   - **Repository permissions**:
     - Issues: Read and write (ラベル操作に必要)
     - Metadata: Read-only

4. **Where can this GitHub App be installed?**
   - Only on this account

5. 作成後:
   - App ID をメモ (`LABEL_APP_ID`)
   - Private key を生成してダウンロード (`LABEL_PRIVATE_KEY`)
   - **Install App** → 自分のアカウント → All repositories を選択

### 2. Railway にデプロイ

1. Railway で新規プロジェクトを作成
2. このリポジトリをデプロイ
3. 環境変数を設定:
   - `GITHUB_APP_ID`: メインAppのID
   - `GITHUB_PRIVATE_KEY`: メインAppの秘密鍵（改行を `\n` に置換）
   - `LABEL_APP_ID`: ラベル操作AppのID
   - `LABEL_PRIVATE_KEY`: ラベル操作Appの秘密鍵（改行を `\n` に置換）
   - `WEBHOOK_SECRET`: Webhook secret

### 3. メインApp をインストール

1. メインAppの設定画面 → Install App
2. 対象のアカウント/Organization を選択
3. All repositories または特定のリポジトリを選択

## 環境変数

| 変数名 | 説明 |
|--------|------|
| `GITHUB_APP_ID` | メインApp（リポジトリ操作用）の ID |
| `GITHUB_PRIVATE_KEY` | メインAppの秘密鍵（PEM形式） |
| `LABEL_APP_ID` | ラベル操作専用Appの ID |
| `LABEL_PRIVATE_KEY` | ラベル操作Appの秘密鍵（PEM形式） |
| `WEBHOOK_SECRET` | Webhook の署名検証用シークレット |
| `PORT` | サーバーポート（Railway が自動設定） |

## Private Key の設定

Railway では改行を含む環境変数の設定が必要です:

```bash
# ファイルの内容を1行に変換
cat private-key.pem | tr '\n' '\\n' | sed 's/\\n$//'
```

または Railway のダッシュボードで複数行入力が可能です。

## ローカル開発

### 方法1: Go で直接実行

```bash
# 依存関係のインストール
go mod download

# .env ファイルを作成
cp .env.example .env
# .env を編集して環境変数を設定

# 実行
go run main.go

# ngrok でトンネル作成（Webhook テスト用）
ngrok http 8080
```

### 方法2: Docker で実行

#### 2-1. .env ファイルを作成

```bash
cp .env.example .env
```

`.env` ファイルを編集して環境変数を設定:

```env
# メインApp（リポジトリ操作用）
GITHUB_APP_ID=123456
GITHUB_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----

# ラベル操作専用App
LABEL_APP_ID=789012
LABEL_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----\nMIIEpQIBAAKCAQEB...\n-----END RSA PRIVATE KEY-----

# Webhook Secret
WEBHOOK_SECRET=your-webhook-secret

# Port
PORT=8080
```

**注意**: 秘密鍵は改行を `\n` に置換してください。

#### 2-2. Docker イメージをビルド

```bash
docker build -t github-setup-app .
```

#### 2-3. コンテナを起動

```bash
docker run -p 8080:8080 --env-file .env github-setup-app
```

または、環境変数を直接指定:

```bash
docker run -p 8080:8080 \
  -e GITHUB_APP_ID=123456 \
  -e GITHUB_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----" \
  -e LABEL_APP_ID=789012 \
  -e LABEL_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----" \
  -e WEBHOOK_SECRET=your-webhook-secret \
  github-setup-app
```

#### 2-4. 動作確認

```bash
# ヘルスチェック
curl http://localhost:8080/health
# レスポンス: OK
```

#### 2-5. ngrok でトンネル作成（Webhook テスト用）

```bash
ngrok http 8080
```

ngrok の URL を GitHub App の Webhook URL に設定:
```
https://xxxx-xx-xxx-xxx-xx.ngrok-free.app/webhook
```

### Docker Compose で実行

`docker-compose.yml` を作成:

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    env_file:
      - .env
    restart: unless-stopped
```

実行:

```bash
# 起動
docker-compose up -d

# ログ確認
docker-compose logs -f

# 停止
docker-compose down
```

## 作成されるラベル

| ラベル | 色 | 説明 |
|--------|-----|------|
| bug | 🔴 | バグ報告 |
| enhancement | 🔵 | 新機能追加 |
| documentation | 🔵 | ドキュメント改善 |
| refactor | 🟡 | リファクタリング |
| performance | 🟣 | パフォーマンス改善 |
| dependencies | 🔵 | 依存関係の更新 |

## 動作フロー

1. 新しいリポジトリが作成される
2. メインAppが `repository.created` イベントを受信
3. リポジトリに `APP_ID` と `APP_PRIVATE_KEY` のシークレットを自動登録（ラベル操作App用）
4. setup-labels ワークフローファイルを自動追加
5. setup-labels ワークフローが実行される
   - ラベル操作Appの認証情報を使用
   - 既存ラベル削除とカスタムラベル作成
6. ワークフローが成功完了すると `workflow_run.completed` イベントを受信
7. メインAppがリポジトリを自動削除

**注意**:
- setup-labels ワークフローが成功完了すると、そのリポジトリは自動的に削除されます。
- シークレットは各リポジトリに自動的に登録されるため、手動での設定は不要です。

## セキュリティ設計

- **権限分離**: リポジトリ操作用とラベル操作用で別々のGitHub Appを使用
- **最小権限**: ラベル操作Appは Issues 権限のみ
- **短命**: リポジトリはワークフロー完了後すぐに削除される
- **秘密鍵の分離**: 各リポジトリに配布される秘密鍵はラベル操作用のみ
