# pt [![Build Status](https://travis-ci.org/hoffa/pt.svg?branch=master)](https://travis-ci.org/hoffa/pt) [![Go Report Card](https://goreportcard.com/badge/github.com/hoffa/pt)](https://goreportcard.com/report/github.com/hoffa/pt)

A minimalist static blog generator.

- Super simple
- Write pages in [Markdown](https://daringfireball.net/projects/markdown/syntax)
- Generates valid [RSS 2.0](https://validator.w3.org/feed/docs/rss2.html) feed

## Installation

```shell
go get github.com/hoffa/pt
```

## Usage

```shell
pt -base-url https://my.site *.md
```

## Front matter

```toml
title = "Hello, world!"
date = 2019-02-11
```

Each page can contain a TOML front matter. It must be placed at the top within `+++` delimiters.

Valid variables are:

- `title`: the content title
- `date`: the content creation date
- `exclude`: if `true`, the page won't be included in `.Pages`

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
title = "Jane Doe"
date = 2019-01-01
exclude = true
+++

Hello, _world_!

This is an example **paragraph**.
EOF
```

And a post:

```shell
cat > my-first-post.md << EOF
+++
title = "My first post"
date = 2019-04-20
+++

This is an example post.

Nothing much to see.
EOF
```

Finally, build:

```shell
pt *.md
```
