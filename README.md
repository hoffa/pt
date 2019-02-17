# pt

[![Build Status](https://travis-ci.org/hoffa/pt.svg?branch=master)](https://travis-ci.org/hoffa/pt) [![Go Report Card](https://goreportcard.com/badge/github.com/hoffa/pt)](https://goreportcard.com/report/github.com/hoffa/pt)

A minimalist static site generator.

Just converts [Markdown](https://daringfireball.net/projects/markdown/syntax) to HTML. Uses Go's [`text/template`](https://golang.org/pkg/text/template/).

## Features

- Tiny
- Straightforward
- Single-file template
- Generates pages instantly
- Generates RSS feed

## Demo

My [website](https://rehn.me). Generated from [hoffa/hoffa.github.io](https://github.com/hoffa/hoffa.github.io).

## Installation

```shell
go get github.com/hoffa/pt
```

## Usage

```shell
pt
```

`pt` takes no arguments.

Markdown files within the working directory are converted to HTML. All pages are generated from [`template.html`](template.html).

## Configuration

Configuration is defined in `pt.toml`.

### Example

```toml
author = "Jane Doe"
email = "jane@doe.com"
baseURL = "https://doe.com"
```

### Variables

- `author`: the site author
- `email`: the author's email
- `baseURL`: the URL used to resolve relative paths

## Front matter

Each page can contain a front matter (the page's metadata).

The front matter is written in [TOML](https://github.com/toml-lang/toml), and must be placed at the top within `+++` delimiters.

### Example

```toml
title = "Hello, world!"
date = 2019-02-11
```

### Variables

- `title`: the content title
- `description`: the content description
- `date`: the content creation date
- `exclude`: if `true`, the page won't be included in `.Site.Pages`
