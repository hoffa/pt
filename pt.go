package main

import (
	"bytes"
	"flag"
	"fmt"
	htmlTemplate "html/template"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	textTemplate "text/template"
	"time"

	"github.com/russross/blackfriday"
	"gopkg.in/yaml.v3"
)

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
	i := bytes.Index(b[3:], delim)
	if !bytes.Equal(b[:3], delim) || i == -1 {
		return nil, b
	}
	return b[3 : i+3], b[i+6:]
}

func parsePage(p, baseURL string) *Page {
	b, err := ioutil.ReadFile(p)
	check(err)
	fm, md := separateContent(b)
	frontMatter := &FrontMatter{Title: p}
	check(yaml.Unmarshal(fm, frontMatter))
	target := replaceExtension(p, ".html")
	return &Page{
		FrontMatter: frontMatter,
		Path:        target,
		URL:         htmlTemplate.URL(urlJoin(baseURL, target)),
		Content:     htmlTemplate.HTML(blackfriday.MarkdownCommon(md)),
	}
}

func writePage(templatePath string, page *Page) {
	tmpl := htmlTemplate.Must(htmlTemplate.ParseFiles(templatePath))
	f, err := os.Create(page.Path)
	check(err)
	defer f.Close()
	check(tmpl.Execute(f, page))
}

func writeRSS(templatePath string, page *Page) {
	tmpl := textTemplate.Must(textTemplate.ParseFiles(templatePath))
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
	flag.Parse()

	var included []*Page
	var excluded []*Page
	for _, p := range flag.Args() {
		page := parsePage(p, *baseURL)
		if page.Exclude {
			excluded = append(excluded, page)
		} else {
			included = append(included, page)
		}
	}
	sort.Slice(included, func(i, j int) bool { return included[i].Date.After(included[j].Date) })
	for _, page := range append(included, excluded...) {
		page.Pages = included
		writePage(*pageTemplatePath, page)
		fmt.Println(page.Path)
	}
	if len(included) > 0 {
		writeRSS(*feedTemplatePath, &Page{
			FrontMatter: &FrontMatter{
				Date: time.Now(),
			},
			Path:  *feedPath,
			URL:   htmlTemplate.URL(urlJoin(*baseURL, *feedPath)),
			Pages: included,
		})
	}
}
