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

type Config struct {
	Author          string
	DateFormat      string
	Email           string
	BaseURL         string
	PreviewImageURL string
}

type FrontMatter struct {
	Title       string
	Description string
	Date        string
}

// Also store a list of all posts??
type Post struct {
	Config      Config
	FrontMatter FrontMatter
	Date        time.Time
	Path        string
	Content     string
	Join        func(base, p string) string
	Posts       []*Post
}

func writeRSS(config *Config, posts []*Post) error {
	author := &feeds.Author{Name: config.Author, Email: config.Email}
	feed := &feeds.Feed{
		Title:  config.Author,
		Link:   &feeds.Link{Href: config.BaseURL},
		Author: author,
	}
	var items []*feeds.Item
	for _, post := range posts {
		if post.FrontMatter.Title == "" {
			continue
		}
		items = append(items, &feeds.Item{
			Title:       post.FrontMatter.Title,
			Author:      author,
			Link:        &feeds.Link{Href: post.Join(config.BaseURL, post.Path)},
			Created:     post.Date,
			Description: post.FrontMatter.Description,
		})
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
		// Assume everything is Markdown
		return nil, b
	}
	return b[3 : i+3], b[i+6:]
}

func executeTemplate(post *Post) error {
	f, err := os.Create(post.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(f, "template", post)
}

func replaceExtension(p, ext string) string {
	return p[:len(p)-len(filepath.Ext(p))] + ext
}

func main() {
	var config Config
	_, err := toml.DecodeFile("pt.toml", &config)
	if err != nil {
		panic(err)
	}
	var posts []*Post
	if err := filepath.Walk(".", func(p string, f os.FileInfo, err error) error {
		if filepath.Ext(p) != ".md" {
			return nil
		}
		b, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		var frontMatter FrontMatter
		fm, md := separateFrontMatter(b)
		if err := toml.Unmarshal(fm, &frontMatter); err != nil {
			return err
		}
		date, err := time.Parse(config.DateFormat, frontMatter.Date)
		if err != nil {
			return err
		}
		fmt.Println(p, frontMatter)
		target := replaceExtension(p, ".html")
		u, err := url.Parse(config.BaseURL)
		if err != nil {
			return err
		}
		u.Path = path.Join(u.Path, target)
		posts = append(posts, &Post{
			Config:      config,
			FrontMatter: frontMatter,
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
	sort.Slice(posts, func(i, j int) bool { return posts[i].Date.After(posts[j].Date) })
	for _, post := range posts {
		post.Posts = posts
		if err := executeTemplate(post); err != nil {
			panic(err)
		}
	}
	if err := writeRSS(&config, posts); err != nil {
		panic(err)
	}
}
