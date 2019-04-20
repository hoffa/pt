package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/feeds"
	"github.com/russross/blackfriday"
)

var config struct {
	rootPath      string
	configPath    string
	rssPath       string
	templatePath  string
	summaryLength int
}

// Site represents the config in pt.toml.
type Site struct {
	Author  string
	BaseURL string
	Params  map[string]interface{}
	Pages   []*Page
}

// FrontMatter represents a page's TOML front matter.
type FrontMatter struct {
	Title   string
	Date    time.Time
	Exclude bool
	Params  map[string]interface{}
}

// Page represents a Markdown page with optional front matter.
// The struct is passed to template.html during template execution.
type Page struct {
	*FrontMatter
	Site    *Site
	Path    string
	Content template.HTML
	Summary template.HTML
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func replaceExtension(p, ext string) string {
	return p[:len(p)-len(filepath.Ext(p))] + ext
}

func joinURL(base, p string) string {
	u, err := url.Parse(base)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join(u.Path, p)
	return u.String()
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

func summarize(s string) string {
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
	return strings.Join(summary, " ")
}

func parsePage(site *Site, p string) *Page {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		panic(err)
	}
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
		Site:        site,
		Path:        replaceExtension(p, ".html"),
		Content:     template.HTML(content),
		Summary:     template.HTML(summarize(content)),
	}
}

func writePage(templatePath string, funcMap template.FuncMap, page *Page) {
	tmpl := template.Must(template.New(templatePath).Funcs(funcMap).ParseFiles(templatePath))
	f, err := os.Create(page.Path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, page); err != nil {
		panic(err)
	}
}

func writeRSS(pages []*Page, site *Site) {
	feed := &feeds.Feed{
		Title:   site.Author,
		Link:    &feeds.Link{Href: site.BaseURL},
		Updated: time.Now(),
	}
	var items []*feeds.Item
	for _, page := range pages {
		items = append(items, &feeds.Item{
			Title:       page.Title,
			Author:      &feeds.Author{Name: site.Author},
			Link:        &feeds.Link{Href: joinURL(site.BaseURL, page.Path)},
			Created:     page.Date,
			Description: string(page.Summary),
			Content:     string(page.Content),
		})
	}
	feed.Items = items
	f, err := os.Create(config.rssPath)
	if err != nil {
		panic(err)
	}
	if err := feed.WriteRss(f); err != nil {
		panic(err)
	}
}

func main() {
	flag.StringVar(&config.configPath, "config", "pt.toml", "config path")
	flag.StringVar(&config.rssPath, "rss", "feed.xml", "RSS feed path")
	flag.StringVar(&config.templatePath, "template", "template.html", "template path")
	flag.IntVar(&config.summaryLength, "summary-length", 150, "summary length in words")
	flag.Parse()
	if flag.NArg() > 0 {
		config.rootPath = flag.Arg(0)
	} else {
		config.rootPath = "."
	}

	var site Site
	_, err := toml.DecodeFile(config.configPath, &site)
	if err != nil {
		fmt.Println("warning:", err)
	}
	var included []*Page
	var excluded []*Page
	if err := filepath.Walk(config.rootPath, func(p string, f os.FileInfo, err error) error {
		if filepath.Ext(p) == ".md" {
			fmt.Println(p)
			page := parsePage(&site, p)
			if page.Exclude {
				excluded = append(excluded, page)
			} else {
				included = append(included, page)
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
	sort.Slice(included, func(i, j int) bool { return included[i].Date.After(included[j].Date) })
	site.Pages = included
	funcMap := template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"absURL": func(p string) string {
			return joinURL(site.BaseURL, p)
		},
		"first": func(n int, v interface{}) []interface{} {
			var l []interface{}
			vv := reflect.ValueOf(v)
			for i := 0; i < min(n, vv.Len()); i++ {
				l = append(l, vv.Index(i).Interface())
			}
			return l
		},
	}
	for _, page := range append(included, excluded...) {
		writePage(config.templatePath, funcMap, page)
	}
	writeRSS(included, &site)
}
