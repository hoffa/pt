package main

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/feeds"
	"github.com/russross/blackfriday"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Author          string
	Email           string
	DateFormat      string
	BaseURL         string
	GoogleAnalytics string
}

type FrontMatter struct {
	Title string
	Date  string
}

type Post struct {
	Title           string
	Date            time.Time
	URL             *url.URL
	Path            string
	Content         string
	GoogleAnalytics string
	Format          func(t time.Time) string
}

type Index struct {
	Posts           []Post
	GoogleAnalytics string
}

func replaceExtension(path, ext string) string {
	return path[0:len(path)-len(filepath.Ext(path))] + ext
}

func executeTemplate(source, target string, data interface{}) error {
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()
	tmpl, err := template.ParseFiles("layout.tmpl", source)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(f, "layout", data)
}

func separateFrontMatter(b []byte) ([]byte, []byte) {
	i := bytes.Index(b[3:], []byte("+++"))
	if i == -1 {
		// Assume everything is Markdown
		return nil, b
	}
	return b[3 : i+3], b[i+6:]
}

func writeRSS(config *Config, posts []Post) error {
	author := &feeds.Author{Name: config.Author, Email: config.Email}
	feed := &feeds.Feed{
		Title:  config.Author,
		Link:   &feeds.Link{Href: config.BaseURL},
		Author: author,
	}
	var items []*feeds.Item
	for _, post := range posts {
		if post.Title == "" {
			continue
		}
		items = append(items, &feeds.Item{
			Title:   post.Title,
			Author:  author,
			Link:    &feeds.Link{Href: post.URL.String()},
			Created: post.Date,
		})
	}
	feed.Items = items
	f, err := os.Create("feed.xml")
	if err != nil {
		return err
	}
	return feed.WriteRss(f)
}

func main() {
	var config Config
	_, err := toml.DecodeFile("pt.toml", &config)
	if err != nil {
		log.Fatal(err)
	}
	base, err := url.Parse(config.BaseURL)
	if err != nil {
		log.Fatal(err)
	}
	index := Index{
		GoogleAnalytics: config.GoogleAnalytics,
	}
	if err := filepath.Walk(".", func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}
		log.Println(path)
		target := replaceExtension(path, ".html")
		u, err := url.Parse(target)
		if err != nil {
			return err
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		var frontMatter FrontMatter
		var date time.Time
		fm, md := separateFrontMatter(b)
		if fm != nil {
			if err := toml.Unmarshal(fm, &frontMatter); err != nil {
				return err
			}
			date, err = time.Parse("2006-01-02", frontMatter.Date)
			if err != nil {
				return err
			}
		} else {
			log.Warn("missing front matter")
		}
		content := string(blackfriday.MarkdownCommon(md))
		// &Post ?
		post := Post{
			Title:           frontMatter.Title,
			Date:            date,
			URL:             base.ResolveReference(u),
			Path:            target,
			Content:         content,
			GoogleAnalytics: config.GoogleAnalytics,
			Format: func(t time.Time) string {
				return t.Format(config.DateFormat)
			},
		}
		index.Posts = append(index.Posts, post)
		return executeTemplate("post.tmpl", target, post)
	}); err != nil {
		log.Fatal(err)
	}
	sort.Slice(index.Posts, func(i, j int) bool { return index.Posts[i].Date.After(index.Posts[j].Date) })
	executeTemplate("index.tmpl", "index.html", index)
	if err := writeRSS(&config, index.Posts); err != nil {
		log.Fatal(err)
	}
}
