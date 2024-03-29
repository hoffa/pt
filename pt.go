package main

import (
	"bytes"
	_ "embed"
	"flag"
	htmlTemplate "html/template"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	textTemplate "text/template"
	"time"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v2"
)

//go:embed templates/page.html
var defaultPageTemplate string

//go:embed templates/feed.xml
var defaultFeedTemplate string

// FrontMatter represents a page's front matter.
type FrontMatter struct {
	Title   string
	Date    time.Time
	Exclude bool
}

// Page represents a Markdown page with optional front matter.
// The struct is passed to template.html during template execution.
type Page struct {
	*FrontMatter
	Path    string
	URL     htmlTemplate.URL
	Content htmlTemplate.HTML
	Pages   []*Page
}

// Renderer is a Blackfriday renderer for Chroma.
type Renderer struct {
	html  *blackfriday.HTMLRenderer
	style string
}

func newRenderer(style string) *Renderer {
	return &Renderer{
		html:  blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{}),
		style: style,
	}
}

func (r *Renderer) RenderHeader(w io.Writer, ast *blackfriday.Node) {}
func (r *Renderer) RenderFooter(w io.Writer, ast *blackfriday.Node) {}
func (r *Renderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	if node.Type == blackfriday.CodeBlock {
		lexer := lexers.Get(string(node.CodeBlockData.Info))
		if lexer == nil {
			lexer = lexers.Fallback
		}
		style := styles.Get(r.style)
		if style == nil {
			style = styles.Fallback
		}
		iterator, err := lexer.Tokenise(nil, string(node.Literal))
		check(err)
		check(html.New().Format(w, style, iterator))
		return blackfriday.GoToNext
	}
	return r.html.RenderNode(w, node, entering)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func replaceExtension(p, ext string) string {
	return p[:len(p)-len(filepath.Ext(p))] + ext
}

// Separates front matter from Markdown
func separateContent(b []byte) ([]byte, []byte) {
	delim := []byte("---")
	if len(b) < 3 {
		return nil, b
	}
	i := bytes.Index(b[3:], delim)
	if !bytes.Equal(b[:3], delim) || i == -1 {
		return nil, b
	}
	return b[3 : i+3], b[i+6:]
}

func parsePage(b []byte, target, baseURL, style string) *Page {
	fm, md := separateContent(b)
	frontMatter := &FrontMatter{
		Date: time.Now(),
	}
	check(yaml.Unmarshal(fm, frontMatter))
	var enabledExtensions blackfriday.Extensions = blackfriday.CommonExtensions | blackfriday.Footnotes
	var content []byte
	if style == "" {
		content = blackfriday.Run(md, blackfriday.WithExtensions(enabledExtensions))
	} else {
		content = blackfriday.Run(md, blackfriday.WithExtensions(enabledExtensions), blackfriday.WithRenderer(newRenderer(style)))
	}
	return &Page{
		FrontMatter: frontMatter,
		Path:        target,
		URL:         htmlTemplate.URL(urlJoin(baseURL, target)),
		Content:     htmlTemplate.HTML(content),
	}
}

func writePage(tmpl htmlTemplate.Template, page *Page) {
	if page.Path == "" {
		check(tmpl.Execute(os.Stdout, page))
	} else {
		check(os.MkdirAll(filepath.Dir(page.Path), 0755))
		f, err := os.Create(page.Path)
		check(err)
		defer f.Close()
		check(tmpl.Execute(f, page))
	}
}

func writeRSS(tmpl textTemplate.Template, page *Page) {
	if page.Path == "" {
		check(tmpl.Execute(os.Stdout, page))
	} else {
		check(os.MkdirAll(filepath.Dir(page.Path), 0755))
		f, err := os.Create(page.Path)
		check(err)
		defer f.Close()
		check(tmpl.Execute(f, page))
	}
}

func urlJoin(base, p string) string {
	u, err := url.Parse(base)
	check(err)
	u.Path = path.Join(u.Path, p)
	return u.String()
}

func main() {
	baseURL := flag.String("base-url", "", "base URL")
	pageTemplatePath := flag.String("template", "", "page template path")
	feedPath := flag.String("feed", "feed.xml", "feed target path")
	feedTemplatePath := flag.String("feed-template", "", "feed template path")
	style := flag.String("highlight", "", "code highlight style")
	dir := flag.String("dir", ".", "where to save generated files")
	flag.Parse()

	var included []*Page
	var excluded []*Page
	ps := flag.Args()
	if len(ps) == 0 {
		b, err := io.ReadAll(os.Stdin)
		check(err)
		page := parsePage(b, "", *baseURL, *style)
		page.FrontMatter.Exclude = true
		excluded = append(excluded, page)
	} else {
		for _, p := range ps {
			b, err := os.ReadFile(p)
			check(err)
			target := filepath.Join(*dir, replaceExtension(p, ".html"))
			page := parsePage(b, target, *baseURL, *style)
			if page.Exclude {
				excluded = append(excluded, page)
			} else {
				included = append(included, page)
			}
		}
	}
	sort.Slice(included, func(i, j int) bool { return included[i].Date.After(included[j].Date) })

	pageTemplate := htmlTemplate.Must(htmlTemplate.New("page").Parse(defaultPageTemplate))
	if *pageTemplatePath != "" {
		pageTemplate = htmlTemplate.Must(htmlTemplate.ParseFiles(*pageTemplatePath))
	}

	feedTemplate := textTemplate.Must(textTemplate.New("feed").Parse(defaultFeedTemplate))
	if *feedTemplatePath != "" {
		feedTemplate = textTemplate.Must(textTemplate.ParseFiles(*feedTemplatePath))
	}

	for _, page := range append(included, excluded...) {
		page.Pages = included
		writePage(*pageTemplate, page)
	}
	if *feedPath != "" && len(included) > 0 {
		writeRSS(*feedTemplate, &Page{
			FrontMatter: &FrontMatter{
				Title: *feedPath,
				Date:  time.Now(),
			},
			Path:  *feedPath,
			URL:   htmlTemplate.URL(urlJoin(*baseURL, *feedPath)),
			Pages: included,
		})
	}
}
