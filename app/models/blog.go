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

type BlogPostVisibility int
type BlogPostVisibilityOption struct {
	ID   int
	Name string
}

const (
	// Update BlogPostVisibility.String() accordingly
	BlogPostVisibilityPublic BlogPostVisibility = iota
	BlogPostVisibilityUnlisted
	BlogPostVisibilityPrivate
	BlogPostVisibilityDirect
	BlogPostVisibilityLimited
)

type BlogPost struct {
	gorm.Model
	ID           string
	UserID       string
	User         User
	Title        string
	Body         string
	Visibility   BlogPostVisibility `gorm:"not null;default:0"`
	CreationDate time.Time
}

func (bpv BlogPostVisibility) String() string {
	return []string{"Public", "Unlisted", "Private", "Direct", "Limited"}[bpv]
}

func BlogPostVisibilityOptions() []BlogPostVisibilityOption {
	return []BlogPostVisibilityOption{
		{ID: int(BlogPostVisibilityPublic), Name: BlogPostVisibilityPublic.String()},
		{ID: int(BlogPostVisibilityUnlisted), Name: BlogPostVisibilityUnlisted.String()},
		{ID: int(BlogPostVisibilityPrivate), Name: BlogPostVisibilityPrivate.String()},
	}
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
