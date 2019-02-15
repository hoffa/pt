# pt

[![Build Status](https://travis-ci.org/hoffa/pt.svg?branch=master)](https://travis-ci.org/hoffa/pt)

## Installation

```shell
go get github.com/hoffa/pt
```

## Usage

Put your Markdown files somewhere and run:

```shell
pt
```

## Front matter

Front matter is in TOML. Valid fields are `title` and `date`.

If there is no `title`, the page isn't displayed on the index page.
