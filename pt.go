package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"io/ioutil"
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

func parsePage(p, baseURL, style string) *Page {
	b, err := ioutil.ReadFile(p)
	check(err)
	fm, md := separateContent(b)
	frontMatter := &FrontMatter{
		Title: p,
		Date:  time.Now(),
	}
	check(yaml.Unmarshal(fm, frontMatter))
	target := replaceExtension(p, ".html")
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
	f, err := os.Create(page.Path)
	check(err)
	defer f.Close()
	check(tmpl.Execute(f, page))
}

func writeRSS(tmpl textTemplate.Template, page *Page) {
	f, err := os.Create(page.Path)
	check(err)
	defer f.Close()
	check(tmpl.Execute(f, page))
}

func urlJoin(base, p string) string {
	u, err := url.Parse(base)
	check(err)
	u.Path = path.Join(u.Path, p)
	return u.String()
}

func main() {
	baseURL := flag.String("base-url", "", "base URL")
	pageTemplatePath := flag.String("template", "templates/page.html", "page template")
	feedPath := flag.String("feed", "feed.xml", "feed target")
	feedTemplatePath := flag.String("feed-template", "templates/feed.xml", "feed template")
	style := flag.String("highlight", "", "code highlight style")
	flag.Parse()

	var included []*Page
	var excluded []*Page
	for _, p := range flag.Args() {
		page := parsePage(p, *baseURL, *style)
		if page.Exclude {
			excluded = append(excluded, page)
		} else {
			included = append(included, page)
		}
	}
	sort.Slice(included, func(i, j int) bool { return included[i].Date.After(included[j].Date) })

	pageTemplate, err := htmlTemplate.ParseFiles(*pageTemplatePath)
	if err != nil {
		fmt.Println("cannot parse", *pageTemplatePath, "using default page template")
		pageTemplate = htmlTemplate.Must(htmlTemplate.New(*pageTemplatePath).Parse(defaultPageTemplate))
	}
	feedTemplate, err := textTemplate.ParseFiles(*feedTemplatePath)
	if err != nil {
		fmt.Println("cannot parse", *feedTemplatePath, "using default feed template")
		feedTemplate = textTemplate.Must(textTemplate.New(*feedTemplatePath).Parse(defaultFeedTemplate))
	}

	for _, page := range append(included, excluded...) {
		page.Pages = included
		writePage(*pageTemplate, page)
		fmt.Println(page.Path)
	}
	if *feedPath != "" {
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
