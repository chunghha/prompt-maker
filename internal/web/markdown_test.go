package web

import (
	"testing"
)

func BenchmarkMarkdownToHTML(b *testing.B) {
	s, err := NewServer(Config{})
	if err != nil {
		b.Fatalf("failed to create server: %v", err)
	}

	md := "This is a **test** of the `markdownToHTML` function."
	for range b.N {
		s.markdownToHTML(md)
	}
}
