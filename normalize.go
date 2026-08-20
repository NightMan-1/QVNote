package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"golang.org/x/net/html"
)

var mdToHTML = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	// Allow raw HTML blocks (restored iframe embeds) to pass through.
	goldmark.WithRendererOptions(gmhtml.WithUnsafe()),
)

// ClearHTML normalizes note HTML via an HTML -> markdown -> HTML roundtrip.
// The markdown step collapses arbitrary markup into a canonical vocabulary
// (headings, lists, tables, code blocks, links, images), then a DOM pass
// fixes tables, heading levels, a duplicated title heading and broken images.
// title is the note title, used for duplicate-heading removal.
// On any internal error the original content is returned unchanged.
func ClearHTML(content string, title string) string {
	md, err := HTMLToMarkdown(content)
	if err != nil {
		checkQuiet(err)
		return content
	}

	var buf bytes.Buffer
	if err := mdToHTML.Convert([]byte(md), &buf); err != nil {
		checkQuiet(err)
		return content
	}

	out, err := postProcessHTML(buf.String(), title)
	if err != nil {
		checkQuiet(err)
		return content
	}
	return out
}

func postProcessHTML(content string, title string) (string, error) {
	doc, err := html.Parse(strings.NewReader("<html><body>" + content + "</body></html>"))
	if err != nil {
		return "", err
	}
	var body *html.Node
	walkNodes(doc, func(n *html.Node) {
		if body == nil && n.Type == html.ElementNode && n.Data == "body" {
			body = n
		}
	})
	if body == nil {
		return "", fmt.Errorf("no body found")
	}

	var toRemove []*html.Node
	titleNorm := normalizeWhitespace(title)

	// Drop the first heading if it duplicates the note title. Titles often
	// carry a site suffix (" / Хабрахабр", " | MegaIndex"), so compare
	// against the title with the suffix cut as well.
	titleVariants := map[string]bool{}
	if titleNorm != "" {
		titleVariants[strings.ToLower(titleNorm)] = true
		for _, sep := range []string{" / ", " | ", " — ", " - "} {
			if idx := strings.LastIndex(titleNorm, sep); idx > 0 {
				titleVariants[strings.ToLower(normalizeWhitespace(titleNorm[:idx]))] = true
			}
		}
	}
	firstHeading := findFirstHeading(body)
	if firstHeading != nil && len(titleVariants) > 0 &&
		titleVariants[strings.ToLower(normalizeWhitespace(nodeText(firstHeading)))] {
		toRemove = append(toRemove, firstHeading)
	}

	// Normalize heading levels: the top content level becomes h2, relative
	// nesting is preserved, levels never skip (h4,h1,h3 -> h2,h2,h3).
	headingStack := []int{}
	removed := func(target *html.Node) bool {
		for _, r := range toRemove {
			if r == target {
				return true
			}
		}
		return false
	}
	walkNodes(body, func(n *html.Node) {
		if n.Type != html.ElementNode || !isHeading(n) || removed(n) {
			return
		}
		level := int(n.Data[1] - '0')
		for len(headingStack) > 0 && headingStack[len(headingStack)-1] > level {
			headingStack = headingStack[:len(headingStack)-1]
		}
		if len(headingStack) == 0 || headingStack[len(headingStack)-1] != level {
			headingStack = append(headingStack, level)
		}
		newLevel := len(headingStack) + 1
		if newLevel > 6 {
			newLevel = 6
		}
		n.Data = fmt.Sprintf("h%d", newLevel)
	})

	// Tables get bootstrap classes; broken images are removed.
	walkNodes(body, func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}
		switch n.Data {
		case "table":
			setAttr(n, "class", "table table-sm")
		case "img":
			src := strings.TrimSpace(getAttr(n, "src"))
			if src == "" || strings.HasPrefix(strings.ToLower(src), "file:") {
				toRemove = append(toRemove, n)
			}
		}
	})

	for _, n := range toRemove {
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
	}

	var out bytes.Buffer
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(&out, c); err != nil {
			return "", err
		}
	}
	return out.String(), nil
}

func walkNodes(n *html.Node, fn func(*html.Node)) {
	fn(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkNodes(c, fn)
	}
}

func isHeading(n *html.Node) bool {
	if len(n.Data) != 2 || n.Data[0] != 'h' {
		return false
	}
	return n.Data[1] >= '1' && n.Data[1] <= '6'
}

func findFirstHeading(root *html.Node) *html.Node {
	var found *html.Node
	walkNodes(root, func(n *html.Node) {
		if found == nil && n.Type == html.ElementNode && isHeading(n) {
			found = n
		}
	})
	return found
}

func nodeText(n *html.Node) string {
	var sb strings.Builder
	walkNodes(n, func(c *html.Node) {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
		}
	})
	return sb.String()
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func setAttr(n *html.Node, key string, val string) {
	for i, a := range n.Attr {
		if a.Key == key {
			n.Attr[i].Val = val
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
}

// normalizeWhitespace trims and collapses any whitespace runs into single spaces.
func normalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
