package parser

import (
	"strings"
	"testing"
)

// TestParseHTML_BasicDocument tests parsing a simple HTML document
func TestParseHTML_BasicDocument(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<head>
	<title>Test Document</title>
</head>
<body>
	<h1>Main Heading</h1>
	<p>This is a paragraph.</p>
	<h2>Subheading</h2>
	<p>Another paragraph.</p>
</body>
</html>
`
	doc, err := ParseHTML(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	if doc.Title != "Test Document" {
		t.Errorf("Expected title 'Test Document', got '%s'", doc.Title)
	}

	if len(doc.Sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(doc.Sections))
	}

	if doc.Sections[0].Heading != "Main Heading" {
		t.Errorf("Expected first heading 'Main Heading', got '%s'", doc.Sections[0].Heading)
	}

	if doc.Sections[0].Level != 1 {
		t.Errorf("Expected first heading level 1, got %d", doc.Sections[0].Level)
	}

	if !strings.Contains(doc.Sections[0].Content, "This is a paragraph.") {
		t.Errorf("Expected first section to contain 'This is a paragraph.', got '%s'", doc.Sections[0].Content)
	}
}

// TestParseHTML_WithCodeBlocks tests parsing HTML with code blocks
func TestParseHTML_WithCodeBlocks(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<body>
	<h1>Code Example</h1>
	<p>Here's some code:</p>
	<pre><code>func main() {
	fmt.Println("Hello")
}</code></pre>
	<p>End of example.</p>
</body>
</html>
`
	doc, err := ParseHTML(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	if len(doc.Sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(doc.Sections))
	}

	content := doc.Sections[0].Content
	if !strings.Contains(content, "func main()") {
		t.Errorf("Expected code block to be preserved, got: %s", content)
	}
}

// TestParseHTML_WithLists tests parsing HTML with lists
func TestParseHTML_WithLists(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<body>
	<h1>List Example</h1>
	<ul>
		<li>Item 1</li>
		<li>Item 2</li>
		<li>Item 3</li>
	</ul>
</body>
</html>
`
	doc, err := ParseHTML(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	content := doc.Sections[0].Content
	if !strings.Contains(content, "Item 1") || !strings.Contains(content, "Item 2") {
		t.Errorf("Expected list items to be preserved, got: %s", content)
	}
}

// TestParseHTML_WithSpecialCharacters tests handling of special characters
func TestParseHTML_WithSpecialCharacters(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<body>
	<h1>Special &amp; Characters</h1>
	<p>Testing &lt;brackets&gt; and &quot;quotes&quot;</p>
</body>
</html>
`
	doc, err := ParseHTML(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	if doc.Sections[0].Heading != "Special & Characters" {
		t.Errorf("Expected heading 'Special & Characters', got '%s'", doc.Sections[0].Heading)
	}

	if !strings.Contains(doc.Sections[0].Content, "<brackets>") {
		t.Errorf("Expected special characters to be decoded, got: %s", doc.Sections[0].Content)
	}
}

// TestParseHTML_EmptyDocument tests parsing an empty document
func TestParseHTML_EmptyDocument(t *testing.T) {
	html := `<!DOCTYPE html><html><body></body></html>`
	doc, err := ParseHTML(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	if doc.Title != "" {
		t.Errorf("Expected empty title, got '%s'", doc.Title)
	}

	if len(doc.Sections) != 0 {
		t.Errorf("Expected 0 sections, got %d", len(doc.Sections))
	}
}

// TestParseHTML_MalformedHTML tests handling of malformed HTML
func TestParseHTML_MalformedHTML(t *testing.T) {
	html := `<html><body><h1>Unclosed heading<p>Paragraph</body></html>`
	doc, err := ParseHTML(strings.NewReader(html))
	// The parser should handle malformed HTML gracefully
	if err != nil {
		t.Fatalf("ParseHTML should handle malformed HTML, got error: %v", err)
	}

	// Should still extract some content
	if len(doc.Sections) == 0 {
		t.Error("Expected parser to extract content from malformed HTML")
	}
}

// TestParseHTML_MultipleHeadingLevels tests parsing with various heading levels
func TestParseHTML_MultipleHeadingLevels(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<body>
	<h1>Level 1</h1>
	<p>Content 1</p>
	<h2>Level 2</h2>
	<p>Content 2</p>
	<h3>Level 3</h3>
	<p>Content 3</p>
	<h4>Level 4</h4>
	<p>Content 4</p>
</body>
</html>
`
	doc, err := ParseHTML(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	if len(doc.Sections) != 4 {
		t.Fatalf("Expected 4 sections, got %d", len(doc.Sections))
	}

	expectedLevels := []int{1, 2, 3, 4}
	for i, level := range expectedLevels {
		if doc.Sections[i].Level != level {
			t.Errorf("Section %d: expected level %d, got %d", i, level, doc.Sections[i].Level)
		}
	}
}

// TestParseHTML_NestedCodeBlocks tests code blocks within pre tags
func TestParseHTML_NestedCodeBlocks(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<body>
	<h1>NATS Example</h1>
	<pre><code class="language-go">package main

import "github.com/nats-io/nats.go"

func main() {
	nc, _ := nats.Connect(nats.DefaultURL)
	nc.Publish("foo", []byte("Hello"))
}</code></pre>
</body>
</html>
`
	doc, err := ParseHTML(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	content := doc.Sections[0].Content
	if !strings.Contains(content, "package main") {
		t.Errorf("Expected code block to contain 'package main', got: %s", content)
	}
	if !strings.Contains(content, "nats.Connect") {
		t.Errorf("Expected code block to contain 'nats.Connect', got: %s", content)
	}
}

// TestParseHTML_OrderedAndUnorderedLists tests both list types
func TestParseHTML_OrderedAndUnorderedLists(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<body>
	<h1>Features</h1>
	<ul>
		<li>Fast messaging</li>
		<li>Lightweight</li>
	</ul>
	<h2>Steps</h2>
	<ol>
		<li>Install NATS</li>
		<li>Configure server</li>
	</ol>
</body>
</html>
`
	doc, err := ParseHTML(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	if len(doc.Sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(doc.Sections))
	}

	// Check unordered list
	if !strings.Contains(doc.Sections[0].Content, "Fast messaging") {
		t.Errorf("Expected first section to contain 'Fast messaging', got: %s", doc.Sections[0].Content)
	}

	// Check ordered list
	if !strings.Contains(doc.Sections[1].Content, "Install NATS") {
		t.Errorf("Expected second section to contain 'Install NATS', got: %s", doc.Sections[1].Content)
	}
}

// TestParseHTML_UTF8Characters tests Unicode/UTF-8 character handling
func TestParseHTML_UTF8Characters(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>NATS 文档</title>
</head>
<body>
	<h1>Übersicht</h1>
	<p>NATS supports émojis 🚀 and special characters: ñ, ü, é</p>
</body>
</html>
`
	doc, err := ParseHTML(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	if doc.Title != "NATS 文档" {
		t.Errorf("Expected title 'NATS 文档', got '%s'", doc.Title)
	}

	if doc.Sections[0].Heading != "Übersicht" {
		t.Errorf("Expected heading 'Übersicht', got '%s'", doc.Sections[0].Heading)
	}

	if !strings.Contains(doc.Sections[0].Content, "🚀") {
		t.Errorf("Expected content to contain emoji, got: %s", doc.Sections[0].Content)
	}
}

// TestParseHTML_NoTitle tests document without title tag
func TestParseHTML_NoTitle(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<body>
	<h1>Content Only</h1>
	<p>No title tag in this document.</p>
</body>
</html>
`
	doc, err := ParseHTML(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	if doc.Title != "" {
		t.Errorf("Expected empty title, got '%s'", doc.Title)
	}

	if len(doc.Sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(doc.Sections))
	}
}

// TestParseHTML_ComplexNesting tests deeply nested HTML structures
func TestParseHTML_ComplexNesting(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<body>
	<h1>Main Topic</h1>
	<div class="content">
		<div class="section">
			<p>First paragraph in nested divs.</p>
			<div class="subsection">
				<p>Second paragraph deeper nested.</p>
			</div>
		</div>
	</div>
</body>
</html>
`
	doc, err := ParseHTML(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	content := doc.Sections[0].Content
	if !strings.Contains(content, "First paragraph") {
		t.Errorf("Expected content to contain 'First paragraph', got: %s", content)
	}
	if !strings.Contains(content, "Second paragraph") {
		t.Errorf("Expected content to contain 'Second paragraph', got: %s", content)
	}
}

// TestParseHTML_MixedContent tests various content types together
func TestParseHTML_MixedContent(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<head><title>NATS Documentation</title></head>
<body>
	<h1>Getting Started</h1>
	<p>NATS is a simple, secure messaging system.</p>
	<h2>Installation</h2>
	<p>Follow these steps:</p>
	<ol>
		<li>Download NATS</li>
		<li>Extract archive</li>
	</ol>
	<h2>Example Code</h2>
	<pre><code>nats-server -p 4222</code></pre>
	<p>That's it!</p>
</body>
</html>
`
	doc, err := ParseHTML(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	if doc.Title != "NATS Documentation" {
		t.Errorf("Expected title 'NATS Documentation', got '%s'", doc.Title)
	}

	if len(doc.Sections) != 3 {
		t.Fatalf("Expected 3 sections, got %d", len(doc.Sections))
	}

	// Verify all sections have content
	for i, section := range doc.Sections {
		if section.Content == "" {
			t.Errorf("Section %d has empty content", i)
		}
	}
}
