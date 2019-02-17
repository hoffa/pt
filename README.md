# pt

[![Build Status](https://travis-ci.org/hoffa/pt.svg?branch=master)](https://travis-ci.org/hoffa/pt)

## Features

- Extremely straightforward
- Tiny
- Accessible

## Installation

```shell
go get github.com/hoffa/pt
```

## Usage

```shell
pt
```

`pt` takes no arguments. Markdown files within the working directory are converted to HTML.

## Front matter

Front matter is in TOML. Valid fields are `title` and `date`.

### Example

```toml
title = "PKI for busy people"
date = 2019-02-11
```

### Variables

- `title`
- `description`
- `date`
- `exclude`
