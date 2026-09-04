package warranty_test

import (
	"testing"

	"warrantykeeper/server/internal/warranty"
)

func TestGuessCategory(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"exact keyword", "מזגן", "מזגן"},
		{"keyword within a sentence", "קניתי מזגן חדש לסלון אתמול", "מזגן"},
		{"multi-word category via a shorter keyword", "מכונת כביסה בוש דגם חדש", "מכונת כביסה"},
		{"keyword distinguishes נייד from נייח", "קניתי מחשב נייד חדש", "מחשב נייד"},
		{"desktop keyword", "קניתי מחשב נייח חדש", "מחשב נייח"},
		{"no keyword present", "קניתי מוצר כלשהו מהחנות", ""},
		{"latin text with no hebrew keywords", "Samsung TV bought yesterday", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := warranty.GuessCategory(tt.text)
			if got != tt.want {
				t.Errorf("GuessCategory(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}
