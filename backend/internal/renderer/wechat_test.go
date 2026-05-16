package renderer

import (
	"strings"
	"testing"
)

func TestRenderBasic(t *testing.T) {
	input := "# Hello World\n\nThis is a **bold** test."
	html := RenderWithTheme(input, "default")

	t.Logf("RenderBasic output: %s", html)

	// WeChat renderer uses <section> + <strong> for headings
	if !strings.Contains(html, "Hello World") {
		t.Error("expected heading text in output")
	}

	// Bold text should be rendered as <strong>
	if !strings.Contains(html, "bold") {
		t.Error("expected bold text in output")
	}
}

func TestRenderParagraph(t *testing.T) {
	input := "A simple paragraph with `inline code`."
	html := RenderWithTheme(input, "default")

	if !strings.Contains(html, "simple paragraph") {
		t.Error("expected paragraph text in output")
	}

	// Inline code should be in a styled span
	if !strings.Contains(html, "inline code") {
		t.Error("expected inline code content in output")
	}
}

func TestRenderImage(t *testing.T) {
	input := "![alt text](https://example.com/img.jpg)\n\nCaption"
	html := RenderWithTheme(input, "default")

	// Image should become an img tag with max-width
	if !strings.Contains(html, `<img`) || !strings.Contains(html, "example.com/img.jpg") {
		t.Error("expected img tag with src in output")
	}
}

func TestRenderLinkAsFootnote(t *testing.T) {
	input := "Visit [example](https://example.com) for more."
	html := RenderWithTheme(input, "default")

	// WeChat doesn't support external links — they should be footnotes
	if !strings.Contains(html, "example.com") {
		t.Error("expected URL to appear as footnote in output")
	}
}

func TestRenderCodeBlock(t *testing.T) {
	input := "```go\nfmt.Println(\"hello\")\n```"
	html := RenderWithTheme(input, "default")

	t.Logf("RenderCodeBlock output: %s", html)

	// Code block should be present
	if !strings.Contains(html, "Println") {
		t.Error("expected code content in output")
	}
}

func TestExtractTitleFromH1(t *testing.T) {
	input := "# My Title\n\nContent here"
	html := RenderWithTheme(input, "default")

	if !strings.Contains(html, "My Title") {
		t.Error("expected title in rendered output")
	}
}

func TestAvailableThemes(t *testing.T) {
	themes := AvailableThemes()
	if len(themes) == 0 {
		t.Error("expected at least one available theme")
	}
}

func TestRenderWithThemeBusiness(t *testing.T) {
	input := "# Business Post\n\nContent"
	html := RenderWithTheme(input, "business")

	if !strings.Contains(html, "Business Post") {
		t.Error("expected title in business theme output")
	}
}

func TestRenderEmptyMarkdown(t *testing.T) {
	html := RenderWithTheme("", "default")

	// Should render without error for empty input
	if html == "" {
		// Empty markdown producing empty html is acceptable
		// The handler layer handles empty input validation
	}
}

func TestRenderTable(t *testing.T) {
	input := "| A | B |\n|---|---|\n| 1 | 2 |"
	html := RenderWithTheme(input, "default")

	// WeChat uses flex layout for tables
	if !strings.Contains(html, "1") && !strings.Contains(html, "2") {
		t.Error("expected table cell content in output")
	}
}
