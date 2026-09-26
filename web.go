package main

import (
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// The site itself stays on GitHub Pages at www. This server only answers the
// bare domain: command line clients get the résumé as text, and everyone else
// is sent on to www exactly as GitHub Pages used to do it.

// Clients that print a response body straight into a terminal.
var terminalClient = regexp.MustCompile(`(?i)^(curl|wget|httpie|xh|fetch|libfetch|aria2|lwp-request)\b`)

func isTerminal(r *http.Request) bool {
	return terminalClient.MatchString(r.UserAgent())
}

func webHandler(www string) http.Handler {
	// 78, not 80: a line that fills the last column wraps twice on some
	// Windows consoles.
	colour := resumeText(renderer(termenv.ANSI256), 78, true)
	plain := resumeText(renderer(termenv.Ascii), 78, true)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "User-Agent")

		if !isTerminal(r) {
			http.Redirect(w, r, strings.TrimSuffix(www, "/")+r.URL.RequestURI(), http.StatusMovedPermanently)
			return
		}

		var body string
		switch strings.TrimSuffix(r.URL.Path, "/") {
		case "", "/resume", "/index.html":
			body = colour
		case "/plain", "/resume.txt":
			body = plain
		default:
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, "\n  404. Nothing lives at "+r.URL.Path+" in the terminal version.\n\n  Try  curl mosambiswas.com  or  ssh mosambiswas.com\n\n")
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=300")
		io.WriteString(w, body)
	})
}

// renderer returns a lipgloss renderer that writes nowhere and always uses the
// given colour profile, for building text that is sent later.
func renderer(profile termenv.Profile) *lipgloss.Renderer {
	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(profile)
	return r
}
