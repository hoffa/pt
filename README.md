![](logo.png)

![](logo.svg)

[![Build Status](https://travis-ci.org/hoffa/pt.svg?branch=master)](https://travis-ci.org/hoffa/pt) [![Go Report Card](https://goreportcard.com/badge/github.com/hoffa/pt)](https://goreportcard.com/report/github.com/hoffa/pt)

A minimalist static site generator.

## Features

- Tiny
- Straighforward
- Single-file template
- Minimal configuration
- Write pages in [Markdown](https://daringfireball.net/projects/markdown/syntax)
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

## Templating

All Markdown files in the working directory are converted to HTML.

All pages are generated from a single `template.html` using Go's [`text/template`](https://golang.org/pkg/text/template/).

Each page is passed a [`Page`](https://github.com/hoffa/pt/blob/5b150b52d5856ecadbab6b5ff1fbcc33f2af832e/pt.go#L38-L46) structure. This allows allows the template to access fields such as `.Title` and `.Site.Author`.

Check out the included [`template.html`](template.html) for example usage.

## Configuration

Configuration is defined in `pt.toml`. It's written in [TOML](https://github.com/toml-lang/toml).

Custom parameters can be defined in the `params` table. They're accessible through `.Site.Params`.

### Example

```toml
author = "Jane Doe"
baseURL = "https://doe.com"

[params]
dateFormat = "January 2, 2006"
```

### Variables

- `author`: the site author
- `baseURL`: the URL used to resolve relative paths

## Front matter

Each page can contain a front matter.

The TOML front matter must be placed at the top within `+++` delimiters.

Custom parameters can be defined in the `params` table. They're accessible through `.Params`.

### Example

```toml
title = "Hello, world!"
date = 2019-02-11
```

### Variables

- `title`: the content title
- `date`: the content creation date
- `exclude`: if `true`, the page won't be included in `.Site.Pages`
