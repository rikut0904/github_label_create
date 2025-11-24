package entity

type Label struct {
	Name        string
	Color       string
	Description string
}

func DefaultLabels() []Label {
	return []Label{
		{Name: "bug", Color: "d73a4a", Description: "バグ報告"},
		{Name: "enhancement", Color: "a2eeef", Description: "新機能追加"},
		{Name: "documentation", Color: "0075ca", Description: "ドキュメント改善"},
		{Name: "refactor", Color: "fbca04", Description: "リファクタリング"},
		{Name: "performance", Color: "5319e7", Description: "パフォーマンス改善"},
		{Name: "dependencies", Color: "0366d6", Description: "依存関係の更新"},
	}
}
