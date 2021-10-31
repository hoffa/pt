package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestStdinMarkdown(t *testing.T) {
	cmd := exec.Command("go", "run", "pt.go")
	cmd.Stdin = strings.NewReader("# Hello!\nSome `cool` _arbitrary_ **Markdown**!")
	actual, err := cmd.Output()
	check(err)
	expected := `<!DOCTYPE html>
<html>
  <head>
    <title></title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
  </head>
  <body>
    <h1>Hello!</h1>

<p>Some <code>cool</code> <em>arbitrary</em> <strong>Markdown</strong>!</p>

  </body>
</html>
`
	if string(actual) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, actual)
	}
}
