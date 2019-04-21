<img src="https://rehn.me/assets/pt.svg" width="25%" alt="pt">

[![Build Status](https://travis-ci.org/hoffa/pt.svg?branch=master)](https://travis-ci.org/hoffa/pt) [![Go Report Card](https://goreportcard.com/badge/github.com/hoffa/pt)](https://goreportcard.com/report/github.com/hoffa/pt)

A minimalist static site generator.

## Features

- Super simple
- Tiny
- Single-file template
- Minimal configuration
- Write pages in [Markdown](https://daringfireball.net/projects/markdown/syntax)
- Generates pages instantly
- Generates RSS feed

## Demo

[https://rehn.me](https://rehn.me)

## Installation

```shell
go get github.com/hoffa/pt
```

## Usage

```shell
pt -base-url https://my.site
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
- `exclude`: if `true`, the page won't be included in `.Site.Pages`

## Templating

All Markdown files in the working directory are converted to HTML. They're generated from `template.html` using [`html/template`](https://golang.org/pkg/html/template/).

Each page is passed a `Page` structure. This allows the template to access fields such as `.Title` and `.Site.Author`.

Check out the included [`template.html`](template.html) for example usage.

### Functions

* `absURL s`: join `baseURL` with `s`
* `first n v`: get `n` first values of `v`

## Themes

Just add your CSS in the `<head>`.

The default template only includes the bare minimum for a navbar.

A tiny touch of CSS can greatly improve readability:

```css
body {
  line-height: 1.5;
  max-width: 40em;
  margin: auto;
  padding: 1em;
}
```
