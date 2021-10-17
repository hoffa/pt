# pt

[![Build Status](https://github.com/hoffa/pt/workflows/.github/workflows/workflow.yml/badge.svg)](https://github.com/hoffa/pt/actions)

A minimalist static blog generator.

- Super simple
- Write pages in [Markdown](https://daringfireball.net/projects/markdown/syntax)
- Generates valid [RSS 2.0](https://validator.w3.org/feed/docs/rss2.html) feed
- Supports code highlighting (using [Chroma](https://github.com/alecthomas/chroma))

## Demo

This README is published at https://hoffa.github.io/pt

## Installation

```shell
go get github.com/hoffa/pt
```

## Usage

```shell
Usage of pt:
  -base-url string
    	base URL
  -feed string
    	feed target (default "feed.xml")
  -feed-template string
    	feed template
  -highlight string
    	code highlight style
  -template string
    	page template
```

## Front matter

Each page can contain a YAML front matter. It must be placed at the top within `---` delimiters:

```markdown
---
title: Hello, world!
date: 2019-02-11
---

This is an example page!
```

Valid variables are:

- `title`: the title
- `date`: the creation date
- `exclude`: if `yes`, the page will be excluded from `.Pages` and the RSS feed

## Example

Create a page template as `template.html`:

```html
<!DOCTYPE html>
<html>
  <head>
    <title>{{ .Title }}</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
  </head>
  <body>
    {{ if eq .Path "index.html" }}
      {{ .Content }}
      <ul>
        {{ range .Pages }}
          <li><a href="{{ .URL }}">{{ .Title }}</a> ({{ .Date.Format "January 2, 2006" }})</li>
        {{ end }}
      </ul>
    {{ else }}
      {{ .Content }}
    {{ end }}
  </body>
</html>
```

Create the index page as `index.md`:

```Markdown
---
title: Jane Doe
exclude: yes
---

Subscribe via [RSS](/feed.xml).
```

And a post within a file called `my-first-post.md`:

````Markdown
---
title: My first post
date: 2019-04-20
---

This is an example **Markdown** _post_.
I like `turtles`.

```python
print("Hello!")
```
````

Finally, build:

```shell
pt -base-url https://mysite.com -template template.html -highlight monokailight *.md
```

See the [Chroma Playground](https://swapoff.org/chroma/playground/) for available syntax highlighting styles.
