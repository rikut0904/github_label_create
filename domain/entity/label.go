package entity

type Label struct {
	Name        string
	Color       string
	Description string
}

func DefaultLabels() []Label {
	return []Label{
		{Name: "bug", Color: "d73a4a", Description: "バグ報告"},
		{Name: "feature", Color: "a2eeef", Description: "新機能追加"},
		{Name: "docs", Color: "0075ca", Description: "ドキュメント改善"},
		{Name: "refactor", Color: "fbca04", Description: "リファクタリング"},
		{Name: "other", Color: "5319e7", Description: "その他"},
	}
}
