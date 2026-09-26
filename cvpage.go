package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/muesli/termenv"
)

// cvPage is the résumé as one file that reads right in two places, so the
// curl trick works on GitHub Pages with no server behind it:
//
//	curl -L mosambiswas.com/cv   coloured text in the terminal
//	mosambiswas.com/cv           a styled page in the browser
//
// The HTML head sits inside an OSC escape sequence (ESC ] ... BEL). Terminals
// swallow OSC strings they do not recognise without printing them, so curl
// shows only the résumé. A browser parses the same bytes as HTML: the head
// loads cv.css, which hides the raw text, and cv.js, which turns the colour
// codes into styled spans. The closing tags are left off on purpose; HTML
// does not need them and a terminal would print them.
//
// The OSC must stay on one line and stay short: some terminals end an OSC at
// a newline, and some cap its length.
func cvPage() string {
	head := strings.Join([]string{
		`<!doctype html><html lang="en">`,
		`<meta charset="utf-8">`,
		`<meta name="viewport" content="width=device-width, initial-scale=1">`,
		`<title>Mosam Biswas · résumé, terminal edition</title>`,
		`<meta name="robots" content="noindex">`,
		`<link rel="canonical" href="https://www.mosambiswas.com/resume.html">`,
		`<link rel="icon" href="/favicon.png">`,
		`<link rel="stylesheet" href="/cv/cv.css">`,
		`<script src="/cv/cv.js" defer></script>`,
		`<noscript><a href="/resume.html">Open the résumé</a></noscript>`,
		`<pre id="cv">`,
	}, "")
	const osc = "\x1b]9999;"
	return osc + head + "\a" + resumeText(renderer(termenv.ANSI256), 78, pageLinks)
}

func writeCV(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(cvPage()), 0o644)
}
