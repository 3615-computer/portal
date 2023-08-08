package models

import (
	"html/template"
	"time"

	"github.com/gomarkdown/markdown"
	mdHtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	"gorm.io/gorm"
)

type Author struct {
	gorm.Model
	ID      string
	Name    string
	NameURL string
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

func (b BlogPost) ToHTML() template.HTML {
	return template.HTML(string(mdToHTML([]byte(b.Body))))
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
