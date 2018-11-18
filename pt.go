package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"gopkg.in/russross/blackfriday.v2"
)

type Post struct {
	Path    string
	Title   string
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
		post := Post{
			Path:    target,
			Title:   string(b[2:bytes.IndexByte(b, byte('\n'))]),
			Content: string(blackfriday.Run(b)),
		}
		posts = append(posts, post)
		tmpl.ExecuteTemplate(f, "post.tmpl", post)
	}
	f, err := os.Create("index.html")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	tmpl.ExecuteTemplate(f, "index.tmpl", posts)
}
