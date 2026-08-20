package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

var iframeRE = regexp.MustCompile(`(?s)<iframe[^>]*>.*?</iframe>|<iframe[^>]*/?>`)

// HTMLToMarkdown converts note HTML content to markdown.
// iframe embeds are extracted before conversion and restored afterwards
// as raw HTML blocks (raw HTML is valid inside markdown).
func HTMLToMarkdown(htmlInput string) (string, error) {
	iframes := map[string]string{}
	htmlInput = iframeRE.ReplaceAllStringFunc(htmlInput, func(s string) string {
		key := fmt.Sprintf("QVNIFRAMEPLACEHOLDER%d", len(iframes))
		iframes[key] = s
		return "<p>" + key + "</p>"
	})

	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
		),
	)
	md, err := conv.ConvertString(htmlInput)
	if err != nil {
		return "", err
	}

	for key, raw := range iframes {
		md = strings.Replace(md, key, "\n\n"+raw+"\n\n", 1)
	}
	return md, nil
}
