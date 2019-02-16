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

// Config represents the config in pt.toml.
type Config struct {
	Author          string
	DateFormat      string
	Email           string
	BaseURL         string
	Lang            string
	PreviewImageURL string
}

// FrontMatter represents a page's TOML front matter.
type FrontMatter struct {
	Title       string
	Description string
	Date        string
	Hide        bool
}

// Page represents a Markdown page with optional front matter.
// The struct is passed to template.html during template execution.
type Page struct {
	Config      *Config
	FrontMatter *FrontMatter
	Date        time.Time
	Path        string
	Content     string
	Join        func(base, p string) string
	Pages       []*Page
}

func writeRSS(config *Config, pages []*Page) error {
	author := &feeds.Author{Name: config.Author, Email: config.Email}
	feed := &feeds.Feed{
		Title:  config.Author,
		Link:   &feeds.Link{Href: config.BaseURL},
		Author: author,
	}
	var items []*feeds.Item
	for _, page := range pages {
		if !page.FrontMatter.Hide {
			items = append(items, &feeds.Item{
				Title:       page.FrontMatter.Title,
				Author:      author,
				Link:        &feeds.Link{Href: page.Join(config.BaseURL, page.Path)},
				Created:     page.Date,
				Description: page.FrontMatter.Description,
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

func separateFrontMatter(b []byte) ([]byte, []byte) {
	i := bytes.Index(b[3:], []byte("+++"))
	if i == -1 {
		return nil, b
	}
	return b[3 : i+3], b[i+6:]
}

func executeTemplate(page *Page) error {
	f, err := os.Create(page.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(f, "template", page)
}

func replaceExtension(p, ext string) string {
	return p[:len(p)-len(filepath.Ext(p))] + ext
}

func parseFrontMatter(b []byte, p string, config *Config) FrontMatter {
	var frontMatter FrontMatter
	if err := toml.Unmarshal(b, &frontMatter); err != nil {
		fmt.Println("warning: cannot parse front matter")
	}
	if frontMatter.Title == "" {
		fmt.Println("warning: missing title; using path")
		frontMatter.Title = p
	}
	if frontMatter.Description == "" {
		fmt.Println("warning: missing description; using title")
		frontMatter.Description = frontMatter.Title
	}
	if frontMatter.Date == "" {
		fmt.Println("warning: missing date; using modification time")
		fileInfo, err := os.Stat(p)
		if err != nil {
			panic(err)
		}
		frontMatter.Date = fileInfo.ModTime().Format(config.DateFormat)
	}
	return frontMatter
}

func main() {
	var config Config
	_, err := toml.DecodeFile("pt.toml", &config)
	if err != nil {
		panic(err)
	}
	var pages []*Page
	if err := filepath.Walk(".", func(p string, f os.FileInfo, err error) error {
		if filepath.Ext(p) != ".md" {
			return nil
		}
		fmt.Println(p)
		b, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		fm, md := separateFrontMatter(b)
		frontMatter := parseFrontMatter(fm, p, &config)
		date, err := time.Parse(config.DateFormat, frontMatter.Date)
		if err != nil {
			fmt.Println("warning:", err)
		}
		target := replaceExtension(p, ".html")
		u, err := url.Parse(config.BaseURL)
		if err != nil {
			return err
		}
		u.Path = path.Join(u.Path, target)
		pages = append(pages, &Page{
			Config:      &config,
			FrontMatter: &frontMatter,
			Date:        date,
			Path:        target,
			Content:     string(blackfriday.MarkdownCommon(md)),
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
	sort.Slice(pages, func(i, j int) bool { return pages[i].Date.After(pages[j].Date) })
	for _, page := range pages {
		page.Pages = pages
		if err := executeTemplate(page); err != nil {
			panic(err)
		}
	}
	if err := writeRSS(&config, pages); err != nil {
		panic(err)
	}
}
