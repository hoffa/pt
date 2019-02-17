# pt

[![Build Status](https://travis-ci.org/hoffa/pt.svg?branch=master)](https://travis-ci.org/hoffa/pt)

A minimalist static site generator.

Just converts [Markdown](https://daringfireball.net/projects/markdown/syntax) to HTML.

## Features

- Tiny
- Straightforward
- Single-file template
- Generates pages instantly
- Geberates RSS feed

## Installation

```shell
go get github.com/hoffa/pt
```

## Usage

```shell
pt
```

`pt` takes no arguments.

Markdown files within the working directory are converted to HTML.

## Demo

My [website](https://rehn.me) is generated from the [hoffa/hoffa.github.io](https://github.com/hoffa/hoffa.github.io) repo.

## Front matter

Front matter is in [TOML](https://github.com/toml-lang/toml).

It must be included before the content within an opening and closing `+++`.

In case of missing values, sensible defaults will be used.

### Example

```toml
title = "Hello, world!"
date = 2019-02-11
```

### Variables

#### title

The content title.

#### description

The content description.

#### date

The content creation date.

#### exclude

If `true`, the page won't be included in `.Site.Pages`.
