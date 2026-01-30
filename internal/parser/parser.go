// Package parser provides HTML parsing functionality for extracting structured
// content from NATS documentation pages, including titles, headings, and code blocks.
package parser

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

// Document represents a parsed HTML documentation page
type Document struct {
	Title    string
	Sections []Section
}

// Section represents a subsection within a document
type Section struct {
	Heading string
	Content string
	Level   int
}

// ParseHTML parses an HTML document and extracts structured content
func ParseHTML(r io.Reader) (*Document, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	result := &Document{
		Sections: make([]Section, 0),
	}

	// Extract title
	result.Title = extractTitle(doc)

	// Extract sections based on headings
	extractSections(doc, result)

	return result, nil
}

// extractTitle finds and returns the document title
func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		return extractText(n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := extractTitle(c); title != "" {
			return title
		}
	}

	return ""
}

// extractSections walks the HTML tree and extracts sections based on headings.
// Each section starts with a heading (h1-h6) and includes all content until the next heading.
// Content includes paragraphs, code blocks, lists, and other elements.
func extractSections(n *html.Node, doc *Document) {
	var currentSection *Section
	var contentBuilder strings.Builder

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			// Check if it's a heading
			if level := getHeadingLevel(node.Data); level > 0 {
				// Save previous section if exists
				if currentSection != nil {
					currentSection.Content = strings.TrimSpace(contentBuilder.String())
					doc.Sections = append(doc.Sections, *currentSection)
				}

				// Start new section
				currentSection = &Section{
					Heading: extractText(node),
					Level:   level,
				}
				contentBuilder.Reset()
				return // Don't process children of heading
			}

			// Handle code blocks - preserve formatting
			if node.Data == "pre" || node.Data == "code" {
				contentBuilder.WriteString(extractText(node))
				contentBuilder.WriteString("\n")
				return // Don't process children separately
			}

			// Handle list items
			if node.Data == "li" {
				contentBuilder.WriteString("• ")
				contentBuilder.WriteString(extractText(node))
				contentBuilder.WriteString("\n")
				return
			}

			// Handle paragraphs and other block elements
			if node.Data == "p" || node.Data == "div" {
				text := extractText(node)
				if text != "" {
					contentBuilder.WriteString(text)
					contentBuilder.WriteString("\n\n")
				}
				return
			}
		}

		// Recursively walk children
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(n)

	// Save last section if exists
	if currentSection != nil {
		currentSection.Content = strings.TrimSpace(contentBuilder.String())
		doc.Sections = append(doc.Sections, *currentSection)
	}
}

// getHeadingLevel returns the heading level (1-6) or 0 if not a heading
func getHeadingLevel(tag string) int {
	switch tag {
	case "h1":
		return 1
	case "h2":
		return 2
	case "h3":
		return 3
	case "h4":
		return 4
	case "h5":
		return 5
	case "h6":
		return 6
	default:
		return 0
	}
}

// extractText recursively extracts all text content from a node and its children
func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var text strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text.WriteString(extractText(c))
	}

	return text.String()
}
