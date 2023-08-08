package models

import (
	"html/template"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	mdHtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	"gorm.io/gorm"
)

type Author struct {
	gorm.Model
	ID          string
	NickName    string
	Name        string
	NickNameURL string
}

type BlogPost struct {
	gorm.Model
	ID           string
	AuthorID     string
	Author       Author
	Title        string
	Body         string
	CreationDate time.Time
}

func (b BlogPost) PreviewHTML() template.HTML {
	bodyTruncated := truncateText(b.Body, 100)
	if len(b.Body) >= 100 {
		bodyTruncated = bodyTruncated + " (...)"
	}
	return template.HTML(string(mdToHTML([]byte(bodyTruncated))))
}

func (b BlogPost) ToHTML() template.HTML {
	return template.HTML(string(mdToHTML([]byte(b.Body))))
}

func truncateText(s string, max int) string {
	if max > len(s) {
		return s
	}
	return s[:strings.LastIndexAny(s[:max], " .,:;-")]
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := mdHtml.CommonFlags | mdHtml.HrefTargetBlank
	opts := mdHtml.RendererOptions{Flags: htmlFlags}
	renderer := mdHtml.NewRenderer(opts)

	return bluemonday.UGCPolicy().SanitizeBytes(markdown.Render(doc, renderer))
}
