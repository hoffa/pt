package main

import (
	"bytes"
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
}

type Post struct {
	FrontMatter
	Path    string
	Content string
}

func replaceExtension(path, ext string) string {
	return path[0:len(path)-len(filepath.Ext(path))] + ext
}

func main() {
	files, err := filepath.Glob("posts/*.md")
	if err != nil {
		log.Fatal(err)
	}
	tmpl, err := template.ParseGlob("*.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	var posts []Post
	for _, path := range files {
		target := replaceExtension(path, ".html")
		log.Println(target)
		f, err := os.Create(target)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		i := bytes.Index(b[3:], []byte("---")) + 3
		var frontMatter FrontMatter
		if err := yaml.Unmarshal(b[3:i], &frontMatter); err != nil {
			log.Fatal(err)
		}
		post := Post{
			FrontMatter: frontMatter,
			Path:        target,
			Content:     string(blackfriday.Run(b[i+3:])),
		}
		posts = append(posts, post)
		tmpl.ExecuteTemplate(f, "post.tmpl", post)
	}
	log.Println("index.html")
	f, err := os.Create("index.html")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	sort.Slice(posts, func(i, j int) bool { return posts[i].Date > posts[j].Date })
	tmpl.ExecuteTemplate(f, "index.tmpl", posts)
}
