# セットアップガイド

このドキュメントでは、GitHub Setup App のセットアップ方法を説明します。

## 前提条件

- GitHubアカウント
- Railway アカウント（または他のホスティングサービス）

## 1. GitHub App を2つ作成

### 1-1. メインApp（リポジトリ操作用）

1. **GitHub にアクセス**
   - Settings → Developer settings → GitHub Apps → **New GitHub App**

2. **基本情報を入力**
   ```
   GitHub App name: Repository Setup App
   Homepage URL: https://your-app.railway.app (後で設定可)
   Webhook URL: https://your-app.railway.app/webhook
   Webhook secret: 任意の文字列（メモしておく）
   ```

3. **Permissions を設定**
   - Repository permissions:
     - **Contents**: Read and write
     - **Secrets**: Read and write
     - **Metadata**: Read-only

4. **Subscribe to events**
   - ✅ Repository
   - ✅ Workflow run

5. **Create GitHub App をクリック**

6. **App ID と秘密鍵を取得**
   - App ID をメモ (`GITHUB_APP_ID`)
   - **Generate a private key** で秘密鍵をダウンロード (`GITHUB_PRIVATE_KEY`)

### 1-2. ラベル操作専用App

1. **GitHub にアクセス**
   - Settings → Developer settings → GitHub Apps → **New GitHub App**

2. **基本情報を入力**
   ```
   GitHub App name: Label Manager
   Homepage URL: https://your-app.railway.app
   Webhook: チェックを外す
   ```

3. **Permissions を設定**
   - Repository permissions:
     - **Issues**: Read and write
     - **Metadata**: Read-only

4. **Where can this GitHub App be installed?**
   - **Only on this account** を選択

5. **Create GitHub App をクリック**

6. **App ID と秘密鍵を取得**
   - App ID をメモ (`LABEL_APP_ID`)
   - **Generate a private key** で秘密鍵をダウンロード (`LABEL_PRIVATE_KEY`)

7. **App をインストール**
   - 左サイドバー → **Install App**
   - 自分のアカウントの **Install** をクリック
   - **All repositories** を選択
   - **Install** をクリック

## 2. Railway にデプロイ

### 2-1. プロジェクトを作成

1. [Railway](https://railway.app) にアクセス
2. **New Project** → **Deploy from GitHub repo**
3. このリポジトリを選択

### 2-2. 環境変数を設定

**Variables** タブで以下を設定:

#### メインApp用
```
GITHUB_APP_ID=123456
GITHUB_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----\nMII...\n-----END RSA PRIVATE KEY-----
```

#### ラベル操作App用
```
LABEL_APP_ID=789012
LABEL_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----\nMII...\n-----END RSA PRIVATE KEY-----
```

#### Webhook Secret
```
WEBHOOK_SECRET=your-webhook-secret
```

**秘密鍵の変換方法:**

macOS/Linux:
```bash
cat private-key.pem | tr '\n' '\\n' | sed 's/\\n$//'
```

または手動で改行を `\n` に置換してください。

### 2-3. デプロイURLを取得

1. Railway で自動的にデプロイが完了
2. **Settings** → **Domains** → **Generate Domain**
3. URLをコピー（例: `https://your-app.railway.app`）

## 3. GitHub App の Webhook URL を更新

### メインApp

1. GitHub → Settings → Developer settings → GitHub Apps → メインApp
2. **Webhook URL** を更新: `https://your-app.railway.app/webhook`
3. **Save changes**

## 4. メインApp をインストール

1. メインApp の設定画面 → 左サイドバー → **Install App**
2. 自分のアカウントの **Install** をクリック
3. **All repositories** または特定のリポジトリを選択
4. **Install** をクリック

## 5. 動作確認

1. **新しいリポジトリを作成**してテスト（空のリポジトリでもOK）
2. 数秒後、以下が自動的に実行されます:
   - シークレットが登録される（APP_ID、APP_PRIVATE_KEY）
   - テンプレートファイルが作成される
     - LICENSE（1番目のコミット）
     - CONTRIBUTING.md（2番目のコミット）
     - .github/workflows/setup-labels.yml（3番目のコミット）
   - ワークフローが自動実行される
   - カスタムラベルが設定される
   - ワークフローファイルが削除される

3. **確認項目**:
   - リポジトリに LICENSE と CONTRIBUTING.md が作成されている
   - Settings → Secrets and variables → Actions にシークレットが登録されている
   - Issues → Labels にカスタムラベルが作成されている
   - `.github/workflows/setup-labels.yml` が削除されている

## トラブルシューティング

### ワークフローが 403 エラーで失敗する

**原因**: ラベル操作Appがインストールされていない

**解決**: ラベル操作App → Install App → All repositories を選択

### シークレット暗号化エラー

**原因**: 秘密鍵の形式が間違っている

**解決**:
- `-----BEGIN RSA PRIVATE KEY-----` で始まることを確認
- 改行が `\n` になっているか確認

### Railway でログを確認

```
Railway ダッシュボード → サービス → Deployments → View Logs
```

正常な場合のログ例:
```
Setting up repository: user/repo
Creating secrets for repository: user/repo
Created APP_ID secret
Created APP_PRIVATE_KEY secret
Creating template files for repository: user/repo
Created all template files
Repository setup completed: user/repo
```

## 次のステップ

- [権限設定の詳細](./permissions.md)
- [開発ガイド](./development.md)
- [アーキテクチャ](./architecture.md)
