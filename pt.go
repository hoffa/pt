package main

import (
	"bytes"
	"flag"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/russross/blackfriday"
)

var config struct {
	baseURL          string
	summaryLength    int
	pagesRootPath    string
	pageTemplatePath string
	feedPath         string
	feedTemplatePath string
}

// FrontMatter represents a page's TOML front matter.
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
	Content template.HTML
	Summary string
	Pages   []*Page
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	delim := []byte("+++")
	i := bytes.Index(b[3:], delim)
	if !bytes.Equal(b[:3], delim) || i == -1 {
		return nil, b
	}
	return b[3 : i+3], b[i+6:]
}

func summarizeHTML(s string) string {
	re := regexp.MustCompile("<[^>]*>")
	fields := strings.Fields(re.ReplaceAllString(s, ""))
	var summary []string
	length := 0
	for _, field := range fields {
		if length > config.summaryLength {
			summary = append(summary, "...")
			break
		}
		summary = append(summary, field)
		length += len(field)
	}
	return html.UnescapeString(strings.Join(summary, " "))
}

func parsePage(pages []*Page, p string) *Page {
	b, err := ioutil.ReadFile(p)
	check(err)
	fm, md := separateContent(b)
	var frontMatter FrontMatter
	if err := toml.Unmarshal(fm, &frontMatter); err != nil {
		fmt.Println("warning:", err)
	}
	if frontMatter.Title == "" {
		fmt.Println("warning: missing title; using path")
		frontMatter.Title = p
	}
	content := string(blackfriday.MarkdownCommon(md))
	return &Page{
		FrontMatter: &frontMatter,
		Pages:       pages,
		Path:        replaceExtension(p, ".html"),
		Content:     template.HTML(content),
		Summary:     summarizeHTML(content),
	}
}

func writePage(templatePath string, funcMap template.FuncMap, page *Page) {
	tmpl := template.Must(template.New(templatePath).Funcs(funcMap).ParseFiles(templatePath))
	f, err := os.Create(page.Path)
	check(err)
	defer f.Close()
	check(tmpl.Execute(f, page))
}

func writeRSS(templatePath, path string, funcMap template.FuncMap, pages []*Page) error {
	writePage(templatePath, funcMap, &Page{
		FrontMatter: &FrontMatter{
			Date: time.Now(),
		},
		Path:  path,
		Pages: pages,
	})
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	header := []byte("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
	return ioutil.WriteFile(path, append(header, b...), 0644)
}

func main() {
	flag.StringVar(&config.baseURL, "base-url", "", "base URL")
	flag.IntVar(&config.summaryLength, "summary-length", 150, "summary length in words")
	flag.StringVar(&config.pageTemplatePath, "page-template", "template.html", "page template path")
	flag.StringVar(&config.pagesRootPath, "pages-root", ".", "pages root directory")
	flag.StringVar(&config.feedPath, "feed", "feed.xml", "feed path")
	flag.StringVar(&config.feedTemplatePath, "feed-template", "rss.xml", "feed template path")
	flag.Parse()

	var included []*Page
	var excluded []*Page
	check(filepath.Walk(config.pagesRootPath, func(p string, f os.FileInfo, err error) error {
		if filepath.Ext(p) == ".md" {
			fmt.Println(p)
			page := parsePage(included, p)
			if page.Exclude {
				excluded = append(excluded, page)
			} else {
				included = append(included, page)
			}
		}
		return nil
	}))
	sort.Slice(included, func(i, j int) bool { return included[i].Date.After(included[j].Date) })
	funcMap := template.FuncMap{
		"absURL": func(p string) string {
			u, err := url.Parse(config.baseURL)
			check(err)
			u.Path = path.Join(u.Path, p)
			return u.String()
		},
	}
	for _, page := range append(included, excluded...) {
		writePage(config.pageTemplatePath, funcMap, page)
	}
	check(writeRSS(config.feedTemplatePath, config.feedPath, funcMap, included))
}
