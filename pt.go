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
	Author     string
	DateFormat string
	BaseURL    string
}

type FrontMatter struct {
	Title       string
	Description string
	Date        string
}

// Also store a list of all posts??
type Post struct {
	Title       string
	Description string
	Date        string
	Path        string
	Content     string
	Posts       []Post
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
		target := replaceExtension(path, ".html")
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
			Title:       frontMatter.Title,
			Description: frontMatter.Description,
			Date:        frontMatter.Date,
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
		executeTemplate("layout.tmpl", post.Path, post)
	}
}
