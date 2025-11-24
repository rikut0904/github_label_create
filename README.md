# GitHub Setup App

リポジトリ作成時にラベルとワークフローを自動設定するGitHub App。

## 機能

- リポジトリ作成時に自動実行
- 既存ラベルを削除し、カスタムラベルを作成
- setup-labels ワークフローファイルを追加

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

### 1. GitHub App を作成

1. GitHub Settings → Developer settings → GitHub Apps → New GitHub App
2. 以下を設定:
   - **App name**: 任意の名前
   - **Homepage URL**: Railway のURL（後で設定可）
   - **Webhook URL**: `https://your-app.railway.app/webhook`
   - **Webhook secret**: 任意の文字列（メモしておく）

3. Permissions:
   - **Repository permissions**:
     - Contents: Read and write
     - Issues: Read and write
     - Metadata: Read-only
   - **Subscribe to events**:
     - Repository

4. 作成後:
   - App ID をメモ
   - Private key を生成してダウンロード

### 2. Railway にデプロイ

1. Railway で新規プロジェクトを作成
2. このリポジトリをデプロイ
3. 環境変数を設定:
   - `GITHUB_APP_ID`: App ID
   - `GITHUB_PRIVATE_KEY`: Private key の内容（改行を `\n` に置換）
   - `WEBHOOK_SECRET`: Webhook secret

### 3. GitHub App をインストール

1. GitHub App の設定画面 → Install App
2. 対象のアカウント/Organization を選択
3. All repositories または特定のリポジトリを選択

## 環境変数

| 変数名 | 説明 |
|--------|------|
| `GITHUB_APP_ID` | GitHub App の ID |
| `GITHUB_PRIVATE_KEY` | Private key（PEM形式） |
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

```bash
# 依存関係のインストール
go mod download

# 実行
go run main.go

# ngrok でトンネル作成（Webhook テスト用）
ngrok http 8080
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
