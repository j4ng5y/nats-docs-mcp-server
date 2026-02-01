// Package parser provides functionality for parsing documentation content
// in various formats including HTML and Markdown.
package parser

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// ParseMarkdown parses markdown content and extracts title and sections
func ParseMarkdown(content []byte, filepath string) (*Document, error) {
	md := goldmark.New()
	source := text.NewReader(content)
	doc := md.Parser().Parse(source)

	// Extract title from frontmatter or first heading
	title := extractTitleFromMarkdown(doc, filepath, content)

	// Extract sections from headings and content
	sections := extractMarkdownSections(doc, content)

	return &Document{
		Title:    title,
		Sections: sections,
	}, nil
}

// extractTitleFromMarkdown extracts the title from markdown content
// Tries in order: first H1 heading, filename, fallback
func extractTitleFromMarkdown(doc ast.Node, filepath string, content []byte) string {
	// First, try to find an H1 heading
	walker := doc.FirstChild()
	for walker != nil {
		if heading, ok := walker.(*ast.Heading); ok && heading.Level == 1 {
			// Found H1, extract text
			text := extractTextFromNode(heading, content)
			if text != "" {
				return text
			}
		}
		walker = walker.NextSibling()
	}

	// If no H1 found, try to extract from frontmatter
	if title := extractFrontmatterTitle(content); title != "" {
		return title
	}

	// Fall back to filename
	if filepath != "" {
		parts := strings.Split(filepath, "/")
		filename := parts[len(parts)-1]
		// Remove .md extension
		if strings.HasSuffix(filename, ".md") {
			filename = filename[:len(filename)-3]
		}
		// Convert underscores/hyphens to spaces and capitalize
		filename = strings.ReplaceAll(filename, "_", " ")
		filename = strings.ReplaceAll(filename, "-", " ")
		filename = strings.Title(filename)
		return filename
	}

	return "Untitled"
}

// extractFrontmatterTitle extracts title from YAML frontmatter
func extractFrontmatterTitle(content []byte) string {
	contentStr := string(content)

	// Check for YAML frontmatter (---\n...title: ...\n---\n)
	if strings.HasPrefix(contentStr, "---") {
		lines := strings.Split(contentStr, "\n")
		if len(lines) > 1 {
			endIdx := -1
			for i := 1; i < len(lines); i++ {
				if strings.TrimSpace(lines[i]) == "---" {
					endIdx = i
					break
				}
			}

			if endIdx > 1 {
				// Parse YAML frontmatter
				for i := 1; i < endIdx; i++ {
					line := strings.TrimSpace(lines[i])
					if strings.HasPrefix(line, "title:") {
						title := strings.TrimPrefix(line, "title:")
						title = strings.TrimSpace(title)
						// Remove quotes if present
						title = strings.Trim(title, "\"'")
						if title != "" {
							return title
						}
					}
				}
			}
		}
	}

	return ""
}

// extractTextFromNode extracts plain text from an AST node and its children
// source is the original markdown byte content needed to extract text from segments
func extractTextFromNode(node ast.Node, source []byte) string {
	var buf bytes.Buffer
	walker := node.FirstChild()
	for walker != nil {
		if text, ok := walker.(*ast.Text); ok {
			// Use source to properly extract text from segment
			if source != nil && len(source) > 0 {
				segment := text.Segment
				// Safely extract segment value
				if segment.Start < len(source) && segment.Stop <= len(source) {
					buf.Write(text.Segment.Value(source))
				}
			}
		} else if link, ok := walker.(*ast.Link); ok {
			// Extract text from links
			buf.WriteString(extractTextFromNode(link, source))
		} else if emph, ok := walker.(*ast.Emphasis); ok {
			// Extract text from emphasis
			buf.WriteString(extractTextFromNode(emph, source))
		}
		walker = walker.NextSibling()
	}
	return buf.String()
}

// extractMarkdownSections extracts sections from markdown headings
func extractMarkdownSections(doc ast.Node, source []byte) []Section {
	var sections []Section
	var currentContent strings.Builder
	var currentHeading string
	var currentLevel int

	walker := doc.FirstChild()
	for walker != nil {
		if heading, ok := walker.(*ast.Heading); ok {
			// Save previous section if it exists
			if currentHeading != "" {
				sections = append(sections, Section{
					Heading: currentHeading,
					Content: strings.TrimSpace(currentContent.String()),
					Level:   currentLevel,
				})
				currentContent.Reset()
			}

			// Start new section
			currentHeading = extractTextFromNode(heading, source)
			currentLevel = heading.Level

		} else if _, ok := walker.(*ast.Heading); !ok && currentHeading != "" {
			// Accumulate content for current section
			content := extractContentFromNode(walker, source)
			if content != "" {
				currentContent.WriteString(content)
				currentContent.WriteString("\n")
			}
		}

		walker = walker.NextSibling()
	}

	// Don't forget the last section
	if currentHeading != "" {
		sections = append(sections, Section{
			Heading: currentHeading,
			Content: strings.TrimSpace(currentContent.String()),
			Level:   currentLevel,
		})
	}

	// If no sections were found (no headings), create a single section with all content
	if len(sections) == 0 {
		allContent := extractContentFromNode(doc, source)
		if allContent != "" {
			sections = append(sections, Section{
				Heading: "Content",
				Content: allContent,
				Level:   1,
			})
		}
	}

	return sections
}

// extractContentFromNode extracts readable content from a node
func extractContentFromNode(node ast.Node, source []byte) string {
	switch n := node.(type) {
	case *ast.Paragraph:
		return extractTextFromNode(n, source)
	case *ast.FencedCodeBlock:
		var buf bytes.Buffer
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			buf.Write(line.Value(source))
		}
		return buf.String()
	case *ast.ListItem:
		var buf bytes.Buffer
		walker := n.FirstChild()
		for walker != nil {
			content := extractContentFromNode(walker, source)
			if content != "" {
				buf.WriteString(content)
				buf.WriteString(" ")
			}
			walker = walker.NextSibling()
		}
		return buf.String()
	case *ast.List:
		var buf bytes.Buffer
		walker := n.FirstChild()
		for walker != nil {
			content := extractContentFromNode(walker, source)
			if content != "" {
				buf.WriteString("• ")
				buf.WriteString(content)
				buf.WriteString("\n")
			}
			walker = walker.NextSibling()
		}
		return buf.String()
	case *ast.Blockquote:
		var buf bytes.Buffer
		walker := n.FirstChild()
		for walker != nil {
			content := extractContentFromNode(walker, source)
			if content != "" {
				buf.WriteString(content)
				buf.WriteString(" ")
			}
			walker = walker.NextSibling()
		}
		return buf.String()
	case *ast.HTMLBlock:
		// Skip HTML blocks
		return ""
	case *ast.Heading:
		// Skip headings (they're handled separately)
		return ""
	case *ast.Document:
		// Process children
		var buf bytes.Buffer
		walker := n.FirstChild()
		for walker != nil {
			content := extractContentFromNode(walker, source)
			if content != "" {
				buf.WriteString(content)
				buf.WriteString("\n")
			}
			walker = walker.NextSibling()
		}
		return strings.TrimSpace(buf.String())
	default:
		// For other node types, try to extract text
		text := extractTextFromNode(n, source)
		if text != "" {
			return text
		}
		return ""
	}
}

// TitleCase converts a string to title case
// This is a simple implementation - a more robust one would use unicode.ToTitle
func titleCase(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - ('a' - 'A')
		}
		return r
	}, s[:1]) + strings.ToLower(s[1:])
}

// NormalizeMarkdown removes extra whitespace and normalizes markdown content
func NormalizeMarkdown(content string) string {
	// Remove extra whitespace
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(strings.TrimSpace(content), " ")
}
