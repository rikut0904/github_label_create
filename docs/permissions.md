# 権限設定ガイド

このドキュメントでは、各GitHub Appに必要な権限と、その理由を説明します。

## メインApp（リポジトリ操作用）

### Repository Permissions

| 権限 | アクセスレベル | 理由 |
|------|--------------|------|
| **Contents** | Read and write | ワークフローファイルの作成と削除に必要 |
| **Secrets** | Read and write | リポジトリにシークレット（APP_ID, APP_PRIVATE_KEY）を登録するため |
| **Metadata** | Read-only | リポジトリの基本情報取得（自動的に付与） |

### Subscribe to Events

| イベント | 理由 |
|---------|------|
| **Repository** | `repository.created` イベントを受信して、新規リポジトリのセットアップを開始 |
| **Workflow run** | `workflow_run.completed` イベントを受信して、ワークフローファイルを削除 |

### API 呼び出し

このAppは以下のGitHub APIを使用します:

1. **シークレット登録**
   - `GET /repos/{owner}/{repo}/actions/secrets/public-key`
   - `PUT /repos/{owner}/{repo}/actions/secrets/{secret_name}`

2. **ファイル作成**
   - `PUT /repos/{owner}/{repo}/contents/{path}`

3. **ファイル削除**
   - `GET /repos/{owner}/{repo}/contents/{path}`
   - `DELETE /repos/{owner}/{repo}/contents/{path}`

---

## ラベル操作専用App

### Repository Permissions

| 権限 | アクセスレベル | 理由 |
|------|--------------|------|
| **Issues** | Read and write | ラベルの削除と作成に必要 |
| **Metadata** | Read-only | リポジトリの基本情報取得（自動的に付与） |

### Subscribe to Events

**なし** - このAppはWebhookを受信しません。ワークフロー内から直接APIを呼び出します。

### API 呼び出し

このAppは以下のGitHub APIをワークフロー内から使用します:

1. **ラベル一覧取得**
   - `gh label list --repo {owner}/{repo}`

2. **ラベル削除**
   - `gh label delete {name} --repo {owner}/{repo}`

3. **ラベル作成**
   - `gh label create {name} --repo {owner}/{repo} --color {color} --description {description}`

---

## セキュリティ設計

### 権限分離の理由

2つのAppに分けることで、以下のセキュリティメリットがあります:

#### 1. 最小権限の原則
- 各Appは必要最小限の権限のみを持つ
- 仮に秘密鍵が漏洩しても、被害を最小限に抑えられる

#### 2. 秘密鍵の分離
- **メインApp**: Railway の環境変数のみに保存（リポジトリには配布しない）
- **ラベル操作App**: 各リポジトリに配布（Issues 権限のみなので安全）

#### 3. 影響範囲の限定
| App | 漏洩時の影響 |
|-----|------------|
| メインApp | ワークフローファイルの作成・削除、シークレット登録 |
| ラベル操作App | ラベルの削除・作成のみ |

---

## 権限の確認方法

### メインApp

```
GitHub → Settings → Developer settings → GitHub Apps → メインApp → Permissions & events
```

以下のように設定されていることを確認:

```
Repository permissions:
  Contents: Read and write
  Secrets: Read and write
  Metadata: Read-only

Subscribe to events:
  ✓ Repository
  ✓ Workflow run
```

### ラベル操作App

```
GitHub → Settings → Developer settings → GitHub Apps → ラベル操作App → Permissions & events
```

以下のように設定されていることを確認:

```
Repository permissions:
  Issues: Read and write
  Metadata: Read-only

Subscribe to events:
  (何もチェックされていない)
```

---

## 権限の変更方法

既存のGitHub Appの権限を変更する場合:

1. GitHub → Settings → Developer settings → GitHub Apps
2. 対象のApp名をクリック
3. 左サイドバー → **Permissions & events**
4. 権限を変更
5. **Save changes** をクリック
6. インストール済みのアカウントに通知が送信される
7. 各アカウントで承認が必要

**注意**: 権限を変更すると、インストール済みのアカウントで再承認が必要になります。

---

## よくある質問

### Q: Administration 権限は必要ですか？

**A: いいえ、不要です。**

以前のバージョンではリポジトリ全体を削除していたため Administration 権限が必要でしたが、現在はワークフローファイルのみを削除するため、Contents 権限だけで十分です。

### Q: なぜ2つのAppが必要なのですか？

**A: セキュリティのためです。**

1つのAppで全てを行うこともできますが、秘密鍵を各リポジトリに配布する必要があります。その場合、強力な権限を持つ秘密鍵が漏洩するリスクがあります。

2つに分けることで、リポジトリに配布する秘密鍵は Issues 権限のみとなり、万が一漏洩してもラベル操作しかできません。

### Q: ワークフロー内で GITHUB_TOKEN ではダメなのですか？

**A: 権限不足でラベル削除ができません。**

デフォルトの `GITHUB_TOKEN` では Issues の write 権限が制限されており、ラベルの削除ができません（403 Forbidden エラー）。そのため、GitHub App トークンが必要です。

---

## 関連ドキュメント

- [セットアップガイド](./setup.md)
- [開発ガイド](./development.md)
- [アーキテクチャ](./architecture.md)
