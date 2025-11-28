package entity

// FileContent はファイルの内容を表すインターフェース
type FileContent interface {
	GetPath() string
	GetContent() string
	GetMessage() string
}

type Workflow struct {
	Path    string
	Content string
	Message string
}

func (w Workflow) GetPath() string    { return w.Path }
func (w Workflow) GetContent() string { return w.Content }
func (w Workflow) GetMessage() string { return w.Message }

type File struct {
	Path    string
	Content string
	Message string
}

func (f File) GetPath() string    { return f.Path }
func (f File) GetContent() string { return f.Content }
func (f File) GetMessage() string { return f.Message }

func DefaultSetupLabelsWorkflow() Workflow {
	return Workflow{
		Path:    ".github/workflows/setup-labels.yml",
		Message: "Add setup-labels workflow",
		Content: `name: setup-labels

on:
  push:
    branches:
      - main

jobs:
  setup-labels:
    runs-on: ubuntu-latest
    steps:
      - name: Generate GitHub App Token
        id: generate-token
        uses: actions/create-github-app-token@v1
        with:
          app-id: ${{ secrets.APP_ID }}
          private-key: ${{ secrets.APP_PRIVATE_KEY }}

      - name: Check if already setup
        id: check
        run: |
          if gh label list --repo ${{ github.repository }} --json name --jq '.[].name' | grep -q "^refactor$"; then
            echo "skip=true" >> $GITHUB_OUTPUT
          else
            echo "skip=false" >> $GITHUB_OUTPUT
          fi
        env:
          GH_TOKEN: ${{ steps.generate-token.outputs.token }}

      - name: Delete all existing labels
        if: steps.check.outputs.skip == 'false'
        run: |
          gh label list --repo ${{ github.repository }} --json name --jq '.[].name' | while read -r label; do
            gh label delete "$label" --repo ${{ github.repository }} --yes
          done
        env:
          GH_TOKEN: ${{ steps.generate-token.outputs.token }}

      - name: Create labels
        if: steps.check.outputs.skip == 'false'
        run: |
          labels=(
            "bug|d73a4a|バグ報告"
            "enhancement|a2eeef|新機能追加"
            "documentation|0075ca|ドキュメント改善"
            "refactor|fbca04|リファクタリング"
            "performance|5319e7|パフォーマンス改善"
            "dependencies|0366d6|依存関係の更新"
          )

          for label in "${labels[@]}"; do
            IFS='|' read -r name color description <<< "$label"
            gh label create "$name" --repo ${{ github.repository }} --color "$color" --description "$description"
          done
        env:
          GH_TOKEN: ${{ steps.generate-token.outputs.token }}
`,
	}
}

func DefaultLicenseFile() File {
	return File{
		Path:    "LICENSE",
		Message: "Add LICENSE file",
		Content: `MIT License

Copyright (c) [2025] [rikut0904]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`,
	}
}

func DefaultContributingFile() File {
	return File{
		Path:    "CONTRIBUTING.md",
		Message: "Add CONTRIBUTING.md file",
		Content: `# コントリビューションガイド

## 開発フロー

1. Issueを作成または既存のIssueを確認
2. ` + "`main`" + `ブランチから作業ブランチを作成
3. 変更を実装
4. Pull Requestを作成

## ブランチ命名規則

` + "```" + `
fix/[機能名]   # 新機能
bug/[修正内容]     # バグ修正
docs/[内容]        # ドキュメント
ref/[内容]    # リファクタリング
` + "```" + `

## コミットメッセージ

` + "```" + `
[種別]/[変更内容]

fix/新機能追加
bug/バグ修正
docs/ドキュメント変更
ref/リファクタリング
test/テスト追加・修正
other/ビルド・設定変更
` + "```" + `

## Pull Request

- 関連するIssue番号を記載
- テンプレートに従って記述
- レビュー前にセルフチェック

## コードスタイル

- プロジェクトの既存コードに合わせる
`,
	}
}
