package main

import (
	"bytes"
	"fmt"
	stdhtml "html"
	"net/url"
	"regexp"
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

	// Tables get bootstrap classes; broken images are removed; iframe styles
	// lose absolute/fixed positioning (saved sticky wrappers like habr's
	// would otherwise pin the embed to the top of the note).
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
		case "iframe":
			cleanIframeStyle(n)
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

// stripLeadingSourceLink detects the "first line is the bare source URL"
// pattern found in old notes: the first block is a single <a> whose visible
// text matches its href (http/https, ignoring query/fragment and trailing
// slashes). The block is removed and the link is returned so it can be
// exposed as the note's source URL instead of cluttering the content.
// On any mismatch the original content is returned with an empty link.
func stripLeadingSourceLink(content string) (string, string) {
	doc, err := html.Parse(strings.NewReader("<html><body>" + content + "</body></html>"))
	if err != nil {
		return content, ""
	}
	var body *html.Node
	walkNodes(doc, func(n *html.Node) {
		if body == nil && n.Type == html.ElementNode && n.Data == "body" {
			body = n
		}
	})
	if body == nil {
		return content, ""
	}

	// First meaningful child of body (skip whitespace-only text).
	var first *html.Node
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode && strings.TrimSpace(c.Data) == "" {
			continue
		}
		first = c
		break
	}
	if first == nil || first.Type != html.ElementNode {
		return content, ""
	}

	// The anchor may sit at top level or inside a wrapping block (p/div).
	var anchor *html.Node
	switch first.Data {
	case "a":
		anchor = first
	case "p", "div":
		walkNodes(first, func(n *html.Node) {
			if anchor == nil && n.Type == html.ElementNode && n.Data == "a" {
				anchor = n
			}
		})
		if anchor == nil {
			return content, ""
		}
		// The block must consist of nothing but that anchor's text.
		if normalizeWhitespace(nodeText(first)) != normalizeWhitespace(nodeText(anchor)) {
			return content, ""
		}
	default:
		return content, ""
	}

	href := strings.TrimSpace(getAttr(anchor, "href"))
	if !strings.HasPrefix(strings.ToLower(href), "http") ||
		!sameURL(normalizeWhitespace(nodeText(anchor)), href) {
		return content, ""
	}

	body.RemoveChild(first)
	var out bytes.Buffer
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(&out, c); err != nil {
			return content, ""
		}
	}
	return out.String(), stripTrackingParams(href)
}

// stripTrackingParams removes known tracking query params (utm_*, fbclid,
// gclid and friends) from a lifted source URL. On any parse problem the
// original string is returned.
func stripTrackingParams(rawurl string) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		return rawurl
	}
	q := u.Query()
	changed := false
	for k := range q {
		kl := strings.ToLower(k)
		if strings.HasPrefix(kl, "utm_") ||
			kl == "fbclid" || kl == "gclid" || kl == "yclid" || kl == "igshid" ||
			kl == "_openstat" || kl == "mc_cid" || kl == "mc_eid" {
			q.Del(k)
			changed = true
		}
	}
	if !changed {
		return rawurl
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// sameURL compares a link's visible text with its href, ignoring case, query
// string, fragment and trailing slashes.
func sameURL(text string, href string) bool {
	strip := func(u string) string {
		if i := strings.IndexAny(u, "?#"); i >= 0 {
			u = u[:i]
		}
		return strings.TrimRight(strings.ToLower(strings.TrimSpace(u)), "/")
	}
	t := strip(text)
	return t != "" && t == strip(href)
}

// cleanIframeStyle drops position-related declarations from an iframe's
// inline style (position, top/right/bottom/left, z-index, margin) — saved
// page fragments often carry sticky/absolute positioning that breaks the
// note layout. If nothing is left, the style attribute is removed entirely.
func cleanIframeStyle(n *html.Node) {
	style := getAttr(n, "style")
	if style == "" {
		return
	}
	drop := map[string]bool{
		"position": true, "top": true, "right": true, "bottom": true, "left": true,
		"z-index": true, "margin": true, "margin-top": true, "margin-right": true,
		"margin-bottom": true, "margin-left": true, "transform": true,
	}
	kept := make([]string, 0)
	for _, decl := range strings.Split(style, ";") {
		decl = strings.TrimSpace(decl)
		if decl == "" {
			continue
		}
		prop := strings.ToLower(strings.TrimSpace(strings.SplitN(decl, ":", 2)[0]))
		if !drop[prop] {
			kept = append(kept, decl)
		}
	}
	if len(kept) == 0 {
		for i, a := range n.Attr {
			if a.Key == "style" {
				n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
				return
			}
		}
		return
	}
	setAttr(n, "style", strings.Join(kept, "; "))
}

// normalizeWhitespace trims and collapses any whitespace runs into single spaces.
func normalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// codepenEmbedRe pulls user/slug out of a CodePen URL
// (codepen.io/<user>/pen|embed|fullembedgrid/<slug>) or of the data-slug-hash
// marker attribute.
var codepenEmbedRe = regexp.MustCompile(`(?i)codepen\.io/(?:([A-Za-z0-9_-]+)/)?(?:pen|embed|fullembedgrid)/([A-Za-z0-9_-]+)`)

// codepenSlugAttrRe matches data-slug-hash with quoted or unquoted values
// (CodePen's own fallback markup uses unquoted attributes).
var codepenSlugAttrRe = regexp.MustCompile(`(?i)data-slug-hash=["']?([A-Za-z0-9_-]+)`)

// RenderCodePenEmbeds rewrites CodePen embeds in stored note HTML into local
// <figure> embeds framed through /api/codepen. It runs only when a note is
// displayed (note.json without raw), so the editor and the import pipeline
// keep the original, unexpanded markup.
//
// Two shapes are recognized:
//   - an <iframe> pointing at codepen.io (e.g. a hand-pasted embed) — its src
//     is rewritten to the proxy;
//   - a <figure> carrying a CodePen pen link (defuddle's rendering of CodePen's
//     <p class="codepen" data-slug-hash=...> marker) — replaced by the embed.
func RenderCodePenEmbeds(content string) string {
	if !strings.Contains(strings.ToLower(content), "codepen") {
		return content
	}
	doc, err := html.Parse(strings.NewReader("<html><body>" + content + "</body></html>"))
	if err != nil {
		return content
	}
	var body *html.Node
	walkNodes(doc, func(n *html.Node) {
		if body == nil && n.Type == html.ElementNode && n.Data == "body" {
			body = n
		}
	})
	if body == nil {
		return content
	}

	var replace []*html.Node
	walkNodes(body, func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}
		switch n.Data {
		case "iframe":
			if src := getAttr(n, "src"); codepenEmbedRe.MatchString(src) {
				replace = append(replace, n)
			}
		case "figure":
			if hasCodePenPenLink(n) {
				replace = append(replace, n)
			}
		}
	})

	for _, n := range replace {
		if n.Parent == nil {
			continue
		}
		user, slug, height, theme, tab := codePenParamsFromNode(n)
		if slug == "" {
			continue
		}
		fig := codePenFigureNode(user, slug, height, theme, tab)
		parent := n.Parent
		parent.InsertBefore(fig, n)
		parent.RemoveChild(n)
	}

	var out bytes.Buffer
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(&out, c); err != nil {
			return content
		}
	}
	return out.String()
}

// hasCodePenPenLink reports whether a figure is defuddle's rendering of a
// CodePen embed: it contains a codepen.io pen/embed link and no image (figures
// wrapping images are left alone).
func hasCodePenPenLink(fig *html.Node) bool {
	found := false
	hasImage := false
	walkNodes(fig, func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}
		if n.Data == "img" {
			hasImage = true
		}
		if n.Data == "a" || n.Data == "iframe" {
			if codepenEmbedRe.MatchString(getAttr(n, "href")) || codepenEmbedRe.MatchString(getAttr(n, "src")) {
				found = true
			}
		}
	})
	return found && !hasImage
}

// codePenParamsFromNode extracts user, slug and display options from an embed
// node (iframe or figure), falling back to defaults.
func codePenParamsFromNode(n *html.Node) (user, slug, height, theme, tab string) {
	height, theme, tab = "480", "light", "result"
	for _, a := range n.Attr {
		v := a.Val
		switch strings.ToLower(a.Key) {
		case "src", "href":
			if m := codepenEmbedRe.FindStringSubmatch(v); m != nil {
				if m[1] != "" {
					user = m[1]
				}
				slug = m[2]
			}
			if q := parseCodePenQuery(v); q != nil {
				if q.Get("height") != "" {
					height = q.Get("height")
				}
				if q.Get("theme-id") != "" {
					theme = q.Get("theme-id")
				}
				if q.Get("default-tab") != "" {
					tab = q.Get("default-tab")
				}
			}
		case "data-slug-hash":
			if slug == "" {
				slug = v
			}
		case "data-user":
			if user == "" {
				user = v
			}
		case "data-height":
			height = v
		case "data-theme-id":
			theme = v
		case "data-default-tab":
			tab = v
		}
	}
	// For a figure wrapper the pen URL lives in a descendant node, not on the
	// figure itself — scan the subtree for it (and for data-slug-hash markers).
	if slug == "" || user == "" {
		walkNodes(n, func(c *html.Node) {
			for _, a := range c.Attr {
				if m := codepenEmbedRe.FindStringSubmatch(a.Val); m != nil {
					if user == "" && m[1] != "" {
						user = m[1]
					}
					if slug == "" {
						slug = m[2]
					}
				}
				if m := codepenSlugAttrRe.FindStringSubmatch(a.Val); m != nil && slug == "" {
					slug = m[1]
				}
			}
		})
	}
	return user, slug, height, theme, tab
}

// parseCodePenQuery returns the query values of a CodePen URL, or nil.
func parseCodePenQuery(raw string) url.Values {
	u, err := url.Parse(raw)
	if err != nil || u.RawQuery == "" {
		return nil
	}
	q, err := url.ParseQuery(stdhtml.UnescapeString(u.RawQuery))
	if err != nil {
		return nil
	}
	return q
}

// codePenFigureNode builds the display <figure> for one pen.
func codePenFigureNode(user, slug, height, theme, tab string) *html.Node {
	if user == "" {
		user = "anon"
	}
	src := "/api/codepen/" + user + "/" + slug +
		"?height=" + height + "&theme-id=" + theme + "&default-tab=" + tab

	fig := &html.Node{Type: html.ElementNode, Data: "figure"}
	fig.Attr = append(fig.Attr, html.Attribute{Key: "class", Val: "codepen-figure"})

	iframe := &html.Node{Type: html.ElementNode, Data: "iframe"}
	iframe.Attr = append(iframe.Attr,
		html.Attribute{Key: "src", Val: src},
		html.Attribute{Key: "height", Val: height},
		html.Attribute{Key: "title", Val: "CodePen Embed"},
		html.Attribute{Key: "frameborder", Val: "0"},
		html.Attribute{Key: "scrolling", Val: "no"},
		html.Attribute{Key: "allowfullscreen", Val: "allowfullscreen"},
		html.Attribute{Key: "loading", Val: "lazy"},
	)
	fig.AppendChild(iframe)

	caption := &html.Node{Type: html.ElementNode, Data: "figcaption"}
	caption.Attr = append(caption.Attr, html.Attribute{Key: "class", Val: "codepen-caption"})
	link := &html.Node{Type: html.ElementNode, Data: "a"}
	link.Attr = append(link.Attr,
		html.Attribute{Key: "href", Val: "https://codepen.io/" + user + "/pen/" + slug},
		html.Attribute{Key: "target", Val: "_blank"},
		html.Attribute{Key: "rel", Val: "noopener"},
	)
	link.AppendChild(&html.Node{Type: html.TextNode, Data: "CodePen"})
	caption.AppendChild(link)

	button := &html.Node{Type: html.ElementNode, Data: "button"}
	button.Attr = append(button.Attr,
		html.Attribute{Key: "type", Val: "button"},
		html.Attribute{Key: "class", Val: "codepen-refresh"},
		html.Attribute{Key: "title", Val: "Обновить"},
		html.Attribute{Key: "data-codepen-refresh", Val: src},
	)
	button.AppendChild(&html.Node{Type: html.TextNode, Data: "↻"})
	caption.AppendChild(button)

	fig.AppendChild(caption)
	return fig
}
