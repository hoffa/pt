package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/russross/blackfriday"
)

type Config struct {
	Author          string
	DateFormat      string
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
	Path        string
	Content     string
	Posts       []Post
}

func separateFrontMatter(b []byte) ([]byte, []byte) {
	i := bytes.Index(b[3:], []byte("+++"))
	if i == -1 {
		// Assume everything is Markdown
		return nil, b
	}
	return b[3 : i+3], b[i+6:]
}

func main() {
	var config Config
	_, err := toml.DecodeFile("pt.toml", &config)
	if err != nil {
		panic(err)
	}
	var posts []Post
	if err := filepath.Walk(".", func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}
		target := path[0:len(path)-len(filepath.Ext(path))] + ".html"
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		var frontMatter FrontMatter
		fm, md := separateFrontMatter(b)
		if err := toml.Unmarshal(fm, &frontMatter); err != nil {
			return err
		}
		content := string(blackfriday.MarkdownCommon(md))
		fmt.Println(path, frontMatter)
		// &Post ?
		post := Post{
			Config:      config,
			FrontMatter: frontMatter,
			Path:        target,
			Content:     content,
		}
		posts = append(posts, post)
		return nil
	}); err != nil {
		panic(err)
	}
	for _, post := range posts {
		post.Posts = posts
		f, err := os.Create(post.Path)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		tmpl, err := template.ParseFiles("template.html")
		if err != nil {
			panic(err)
		}
		if err := tmpl.ExecuteTemplate(f, "template", post); err != nil {
			panic(err)
		}
	}
}
