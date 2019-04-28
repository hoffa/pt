<img src="https://rehn.me/assets/pt.svg" width="25%" alt="pt">

[![Build Status](https://travis-ci.org/hoffa/pt.svg?branch=master)](https://travis-ci.org/hoffa/pt) [![Go Report Card](https://goreportcard.com/badge/github.com/hoffa/pt)](https://goreportcard.com/report/github.com/hoffa/pt)

A minimalist static site generator.

## Features

- Super simple
- Tiny
- Write pages in [Markdown](https://daringfireball.net/projects/markdown/syntax)
- Generates pages instantly
- Generates valid [RSS 2.0](https://validator.w3.org/feed/docs/rss2.html) feed

## Demo

[https://rehn.me](https://rehn.me)

## Installation

```shell
go get github.com/hoffa/pt
```

## Usage

```shell
pt -base-url https://my.site *.md
```

## Front matter

Each page can contain a TOML front matter. It must be placed at the top within `+++` delimiters.

### Example

```toml
title = "Hello, world!"
date = 2019-02-11
```

### Variables

- `title`: the content title
- `date`: the content creation date
- `exclude`: if `true`, the page won't be included in `.Pages`

## Templating

All Markdown files in the working directory are converted to HTML. They're generated from [`template.html`](template.html) using [`html/template`](https://golang.org/pkg/html/template/).

Each page is passed a `Page` structure. This allows the template to access fields such as `.Title` and `.URL`.

## Themes

Just add your CSS in the `<head>`.
For example:

```css
body {
  line-height: 1.5;
  max-width: 40em;
  margin: auto;
  padding: 1em;
}
```

## Example

First, get the page and RSS feed templates:

```shell
curl -L https://github.com/hoffa/pt/archive/master.tar.gz \
  | tar xz --strip-components=1 pt-master/templates
```

Now let's create our index page:

```shell
cat > index.md << EOF
+++
title = "Hello!"
date = 2019-01-01
exclude = true
+++

Hello, _world_!

This is an example **paragraph**.
EOF
```

And a post:

```shell
+++
title = "My first post"
date = 2019-04-20
+++

This is an example post.

Nothing much to see.
```

And finally build:

```shell
pt *.md
```
