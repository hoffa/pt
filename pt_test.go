package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestStdinMarkdown(t *testing.T) {
	cmd := exec.Command("go", "run", "pt.go", "-highlight", "monokai")
	cmd.Stdin = strings.NewReader("# Hello!\nSome `cool` _arbitrary_ **Markdown**!\n\n```python\nprint(\"Hi!\")\n```")
	actual, err := cmd.Output()
	check(err)
	expected := `<!DOCTYPE html>
<html lang="">
  <head>
    <title></title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
  </head>
  <body>
    <h1>Hello!</h1>

<p>Some <code>cool</code> <em>arbitrary</em> <strong>Markdown</strong>!</p>
<pre tabindex="0" style="color:#f8f8f2;background-color:#272822;"><code><span style="display:flex;"><span>print(<span style="color:#e6db74"></span><span style="color:#e6db74">&#34;</span><span style="color:#e6db74">Hi!</span><span style="color:#e6db74">&#34;</span>)
</span></span></code></pre>
  </body>
</html>
`
	if string(actual) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, actual)
	}
}
