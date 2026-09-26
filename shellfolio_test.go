package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func get(t *testing.T, path, agent string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("User-Agent", agent)
	rec := httptest.NewRecorder()
	webHandler("https://www.mosambiswas.com").ServeHTTP(rec, req)
	return rec
}

func TestTerminalsGetTheResume(t *testing.T) {
	for _, agent := range []string{"curl/8.9.1", "Wget/1.21.4", "HTTPie/3.2.2", "xh/0.22.0"} {
		rec := get(t, "/", agent)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d", agent, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Sedna Technologies") {
			t.Fatalf("%s: body is not the résumé", agent)
		}
		if !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/plain") {
			t.Fatalf("%s: content type %q", agent, rec.Header().Get("Content-Type"))
		}
	}
}

func TestPlainHasNoEscapes(t *testing.T) {
	body := get(t, "/plain", "curl/8.9.1").Body.String()
	if strings.Contains(body, "\x1b[") {
		t.Fatal("/plain contains ANSI escapes")
	}
	if !strings.Contains(get(t, "/", "curl/8.9.1").Body.String(), "\x1b[") {
		t.Fatal("/ should be coloured")
	}
}

// Browsers must land exactly where GitHub Pages used to send them, path and
// query intact, or old links to the bare domain break.
func TestBrowsersAreRedirectedToWWW(t *testing.T) {
	agent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/140.0 Safari/537.36"
	for path, want := range map[string]string{
		"/":                         "https://www.mosambiswas.com/",
		"/resume.html":              "https://www.mosambiswas.com/resume.html",
		"/PixelShift/?src=card":     "https://www.mosambiswas.com/PixelShift/?src=card",
		"/sheichobi/sheichobi.html": "https://www.mosambiswas.com/sheichobi/sheichobi.html",
	} {
		rec := get(t, path, agent)
		if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != want {
			t.Fatalf("%s: got %d %q, want 301 %q", path, rec.Code, rec.Header().Get("Location"), want)
		}
	}
}

func TestUnknownTerminalPathIs404(t *testing.T) {
	if rec := get(t, "/wp-admin", "curl/8.9.1"); rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
}

// Every line has to fit the width it was rendered for, or it wraps in the
// visitor's terminal and the layout falls apart.
func TestResumeFitsItsWidth(t *testing.T) {
	for _, width := range []int{40, 60, 78} {
		text := resumeText(renderer(termenv.ANSI256), width, true)
		for i, line := range strings.Split(text, "\n") {
			if w := ansi.StringWidth(line); w > width {
				t.Fatalf("width %d, line %d is %d wide: %q", width, i+1, w, ansi.Strip(line))
			}
		}
	}
}

// The same for the SSH app, on every tab, from the smallest window it
// accepts up past the width where it stops growing.
func TestAppFitsTheWindow(t *testing.T) {
	for _, width := range []int{minWidth, 50, 60, 75, 76, 90, 100, 140} {
		for tab := range tabs {
			for _, open := range []bool{false, true} {
				m := newModel(renderer(termenv.ANSI256), func(string) {})
				m.width, m.height = width, 30
				m = m.switchTo(tab)
				m.cursor, m.open = 6, open
				view := m.View()
				lines := strings.Split(view, "\n")
				if len(lines) > m.height {
					t.Fatalf("width %d tab %d: %d lines for a %d line window", width, tab, len(lines), m.height)
				}
				for i, line := range lines {
					if w := ansi.StringWidth(line); w > width {
						t.Fatalf("width %d tab %d open %v, line %d is %d wide: %q", width, tab, open, i+1, w, ansi.Strip(line))
					}
				}
			}
		}
	}
}

// Standing rule for everything Mosam publishes: no em or en dashes.
func TestNoDashesInCopy(t *testing.T) {
	text := resumeText(renderer(termenv.Ascii), 78, true)
	m := newModel(renderer(termenv.Ascii), func(string) {})
	text += m.aboutPage(80) + m.contactPage(80)
	for _, p := range portfolio {
		text += m.detail(p, 60)
	}
	if i := strings.IndexAny(text, "—–"); i >= 0 {
		t.Fatalf("dash found near %q", text[max(i-30, 0):min(i+30, len(text))])
	}
}
