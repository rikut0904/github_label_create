# GitHub Setup App

リポジトリ作成時にラベルとワークフローを自動設定するGitHub App。

## 概要

新しいGitHubリポジトリを作成すると、自動的に:
- カスタムラベルを設定
- ワークフローファイルを追加・実行
- 完了後、ワークフローファイルを自動削除

## 主な機能

✅ **自動ラベル設定**
- 既存ラベルを削除
- カスタムラベル（bug, enhancement, documentation, refactor, performance, dependencies）を作成

✅ **セキュアな設計**
- 権限分離（2つのGitHub Appを使用）
- 最小権限の原則
- 秘密鍵の安全な管理

✅ **完全自動化**
- ユーザー操作不要
- リポジトリ作成するだけで全て完了

## クイックスタート

### 1. セットアップ

詳細な手順は [docs/setup.md](./docs/setup.md) を参照してください。

1. GitHub App を2つ作成（メインApp、ラベル操作App）
2. Railway にデプロイ
3. GitHub App をインストール

### 2. 使用方法

新しいリポジトリを作成するだけです！

```
1. GitHub で新しいリポジトリを作成
   ↓
2. 自動的にラベルが設定される
   ↓
3. 完了（ワークフローファイルは自動削除）
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

## アーキテクチャ

クリーンアーキテクチャに基づいた設計:

```
interface/ (Webhook受信)
    ↓
usecase/ (ビジネスロジック)
    ↓
domain/ (ビジネスルール)
    ↑
infrastructure/ (GitHub API)
```

詳細は [docs/architecture.md](./docs/architecture.md) を参照してください。

## 技術スタック

- **言語**: Go 1.24
- **フレームワーク**: 標準ライブラリ (net/http)
- **デプロイ**: Railway, Docker
- **外部API**: GitHub REST API

## ドキュメント

- [📘 セットアップガイド](./docs/setup.md) - 初期設定の手順
- [🔐 権限設定](./docs/permissions.md) - 必要な権限の詳細
- [💻 開発ガイド](./docs/development.md) - ローカル開発の方法
- [🏗️ アーキテクチャ](./docs/architecture.md) - システム設計

## 必要な権限

### メインApp（リポジトリ操作用）
- Contents: Read and write
- Secrets: Read and write
- Metadata: Read-only

### ラベル操作専用App
- Issues: Read and write
- Metadata: Read-only

詳細は [docs/permissions.md](./docs/permissions.md) を参照してください。

## 環境変数

| 変数名 | 説明 |
|--------|------|
| `GITHUB_APP_ID` | メインApp の ID |
| `GITHUB_PRIVATE_KEY` | メインApp の秘密鍵 |
| `LABEL_APP_ID` | ラベル操作App の ID |
| `LABEL_PRIVATE_KEY` | ラベル操作App の秘密鍵 |
| `WEBHOOK_SECRET` | Webhook の署名検証用 |
| `PORT` | サーバーポート（デフォルト: 8080） |

## ローカル開発

```bash
# 依存関係のインストール
go mod download

# 環境変数を設定
cp .env.example .env
# .env を編集

# 実行
go run main.go

# または Docker で
docker-compose up
```

詳細は [docs/development.md](./docs/development.md) を参照してください。

## セキュリティ

- ✅ 権限分離（2つのGitHub Appを使用）
- ✅ 最小権限の原則
- ✅ Webhook 署名検証
- ✅ シークレット暗号化（libsodium互換）

## ライセンス

MIT License

## 貢献

Issue や Pull Request を歓迎します！

## サポート

問題が発生した場合:
1. [docs/setup.md](./docs/setup.md) のトラブルシューティングを確認
2. [Issues](https://github.com/yourusername/github-setup-app/issues) で報告
