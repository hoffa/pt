# pt

[![Build Status](https://github.com/hoffa/pt/workflows/.github/workflows/workflow.yml/badge.svg)](https://github.com/hoffa/pt/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/hoffa/pt)](https://goreportcard.com/report/github.com/hoffa/pt)

A minimalist static blog generator.

- Super simple
- Write pages in [Markdown](https://daringfireball.net/projects/markdown/syntax)
- Generates valid [RSS 2.0](https://validator.w3.org/feed/docs/rss2.html) feed
- Supports code highlighting (using [Chroma](https://github.com/alecthomas/chroma))

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
    	feed template (default "templates/feed.xml")
  -highlight string
    	code highlight style
  -template string
    	page template (default "templates/page.html")
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

First, get the page and RSS feed templates:

```shell
curl -L https://github.com/hoffa/pt/archive/master.tar.gz \
  | tar -zxf- --strip-components=1 pt-master/templates
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
pt -base-url https://mysite.com -highlight monokailight *.md
```

See the [Chroma Playground](https://swapoff.org/chroma/playground/) for available syntax highlighting styles.
