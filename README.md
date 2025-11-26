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
