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
	Email   string
	BaseURL string
	Params  map[string]interface{}
	Pages   []*Page
}

// FrontMatter represents a page's TOML front matter.
type FrontMatter struct {
	Title       string
	Description string
	Date        time.Time
	Article     bool
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
	i := bytes.Index(b[3:], []byte("+++"))
	if i == -1 {
		return nil, b
	}
	return b[3 : i+3], b[i+6:]
}

func parsePage(p string) (*FrontMatter, string) {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		panic(err)
	}
	fm, md := separateContent(b)
	frontMatter := FrontMatter{Article: true}
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
	return &frontMatter, string(blackfriday.MarkdownCommon(md))
}

func writePages(pages []*Page) error {
	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		return err
	}
	sort.Slice(pages, func(i, j int) bool { return pages[i].Date.After(pages[j].Date) })
	for _, page := range pages {
		f, err := os.Create(page.Path)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := tmpl.ExecuteTemplate(f, "template", page); err != nil {
			return err
		}
	}
	return nil
}

func writeRSS(pages []*Page, site *Site) error {
	author := &feeds.Author{Name: site.Author, Email: site.Email}
	feed := &feeds.Feed{
		Title:   site.Author,
		Link:    &feeds.Link{Href: site.BaseURL},
		Author:  author,
		Updated: time.Now(),
	}
	var items []*feeds.Item
	for _, page := range pages {
		if page.Article {
			items = append(items, &feeds.Item{
				Title:       page.Title,
				Author:      author,
				Link:        &feeds.Link{Href: page.Join(site.BaseURL, page.Path)},
				Created:     page.Date,
				Description: page.Description,
				Content:     page.Content,
			})
		}
	}
	feed.Items = items
	f, err := os.Create("feed.xml")
	if err != nil {
		return err
	}
	return feed.WriteRss(f)
}

func main() {
	var site Site
	_, err := toml.DecodeFile("pt.toml", &site)
	if err != nil {
		panic(err)
	}
	var pages []*Page
	if err := filepath.Walk(".", func(p string, f os.FileInfo, err error) error {
		if filepath.Ext(p) != ".md" {
			return nil
		}
		fmt.Println(p)
		frontMatter, content := parsePage(p)
		pages = append(pages, &Page{
			FrontMatter: frontMatter,
			Site:        &site,
			Path:        replaceExtension(p, ".html"),
			Content:     content,
			Join: func(base, p string) string {
				u, _ := url.Parse(base)
				u.Path = path.Join(u.Path, p)
				return u.String()
			},
		})
		return nil
	}); err != nil {
		panic(err)
	}
	site.Pages = pages
	if err := writePages(pages); err != nil {
		panic(err)
	}
	if err := writeRSS(pages, &site); err != nil {
		panic(err)
	}
}
