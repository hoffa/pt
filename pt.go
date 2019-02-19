package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/feeds"
	"github.com/russross/blackfriday"
)

const (
	configPath    = "pt.toml"
	rssPath       = "feed.xml"
	templatePath  = "template.html"
	summaryLength = 150
)

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
	Content string
	Summary string
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
		if length > summaryLength {
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
	if frontMatter.Date.IsZero() {
		fmt.Println("warning: missing date; using modification time")
		fileInfo, err := os.Stat(p)
		if err != nil {
			panic(err)
		}
		frontMatter.Date = fileInfo.ModTime()
	}
	content := string(blackfriday.MarkdownCommon(md))
	return &Page{
		FrontMatter: &frontMatter,
		Site:        site,
		Path:        replaceExtension(p, ".html"),
		Content:     content,
		Summary:     summarize(content),
	}
}

func writePages(tmpl *template.Template, pages []*Page) {
	for _, page := range pages {
		f, err := os.Create(page.Path)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if err := tmpl.Execute(f, page); err != nil {
			panic(err)
		}
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
			Description: page.Summary,
			Content:     page.Content,
		})
	}
	feed.Items = items
	f, err := os.Create(rssPath)
	if err != nil {
		panic(err)
	}
	if err := feed.WriteRss(f); err != nil {
		panic(err)
	}
}

func main() {
	var site Site
	_, err := toml.DecodeFile(configPath, &site)
	if err != nil {
		fmt.Println("warning:", err)
	}
	var included []*Page
	var excluded []*Page
	if err := filepath.Walk(".", func(p string, f os.FileInfo, err error) error {
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
		"absURL": func(p string) string {
			return joinURL(site.BaseURL, p)
		},
	}
	tmpl := template.Must(template.New(templatePath).Funcs(funcMap).ParseFiles(templatePath))
	writePages(tmpl, included)
	writePages(tmpl, excluded)
	writeRSS(included, &site)
}
