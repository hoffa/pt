package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"text/template"

	"gopkg.in/russross/blackfriday.v2"
	"gopkg.in/yaml.v2"
)

type FrontMatter struct {
	Title string `yaml:"title"`
	Date  string `yaml:"date"`
	Hide  bool   `yaml:"hide"`
}

type Post struct {
	FrontMatter
	Path    string
	Content string
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

func separateFrontMatter(b []byte) ([]byte, []byte, error) {
	i := bytes.Index(b[3:], []byte("---"))
	if i == -1 {
		return nil, nil, errors.New("missing front matter")
	}
	return b[3 : i+3], b[i+6:], nil
}

func main() {
	var posts []Post
	if err := filepath.Walk(".", func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}
		log.Println(path)
		target := replaceExtension(path, ".html")
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		fm, md, err := separateFrontMatter(b)
		if err != nil {
			log.Print(err)
			return nil
		}
		var frontMatter FrontMatter
		if err := yaml.Unmarshal(fm, &frontMatter); err != nil {
			return err
		}
		content := string(blackfriday.Run(md))
		post := Post{
			FrontMatter: frontMatter,
			Path:        target,
			Content:     content,
		}
		posts = append(posts, post)
		return executeTemplate("post.tmpl", target, post)
	}); err != nil {
		log.Fatal(err)
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].Date > posts[j].Date })
	executeTemplate("index.tmpl", "index.html", posts)
}
