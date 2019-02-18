package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/feeds"
	"github.com/russross/blackfriday"
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
	Title       string
	Description string
	Date        time.Time
	Exclude     bool
	Params      map[string]interface{}
}

// Page represents a Markdown page with optional front matter.
// The struct is passed to template.html during template execution.
type Page struct {
	*FrontMatter
	Site    *Site
	Path    string
	Content string
	Join    func(base, p string) string
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
	if frontMatter.Description == "" {
		fmt.Println("warning: missing description; using title")
		frontMatter.Description = frontMatter.Title
	}
	if frontMatter.Date.IsZero() {
		fmt.Println("warning: missing date; using modification time")
		fileInfo, err := os.Stat(p)
		if err != nil {
			panic(err)
		}
		frontMatter.Date = fileInfo.ModTime()
	}
	return &Page{
		FrontMatter: &frontMatter,
		Site:        site,
		Path:        replaceExtension(p, ".html"),
		Content:     string(blackfriday.MarkdownCommon(md)),
		Join: func(base, p string) string {
			u, err := url.Parse(base)
			if err != nil {
				panic(err)
			}
			u.Path = path.Join(u.Path, p)
			return u.String()
		},
	}
}

func writePages(tmpl *template.Template, pages []*Page) {
	for _, page := range pages {
		f, err := os.Create(page.Path)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if err := tmpl.ExecuteTemplate(f, "template", page); err != nil {
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
			Link:        &feeds.Link{Href: page.Join(site.BaseURL, page.Path)},
			Created:     page.Date,
			Description: page.Description,
			Content:     page.Content,
		})
	}
	feed.Items = items
	f, err := os.Create("feed.xml")
	if err != nil {
		panic(err)
	}
	if err := feed.WriteRss(f); err != nil {
		panic(err)
	}
}

func main() {
	var site Site
	_, err := toml.DecodeFile("pt.toml", &site)
	if err != nil {
		fmt.Println("warning:", err)
	}
	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		panic(err)
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
	writePages(tmpl, included)
	writePages(tmpl, excluded)
	writeRSS(included, &site)
}
