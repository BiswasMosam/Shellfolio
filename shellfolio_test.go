package main

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

// The site's text files live on www. A terminal asking the bare domain for one
// must be sent there, not handed the terminal version's 404.
func TestTerminalsAreSentToTheSitesTextFiles(t *testing.T) {
	for _, path := range []string{"/llms.txt", "/llms-full.txt", "/humans.txt", "/robots.txt"} {
		want := "https://www.mosambiswas.com" + path
		rec := get(t, path, "curl/8.9.1")
		if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != want {
			t.Fatalf("%s: got %d %q, want 301 %q", path, rec.Code, rec.Header().Get("Location"), want)
		}
		// Without -L, curl prints this body and stops.
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("%s: body does not say where the file is: %q", path, rec.Body.String())
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
		text := resumeText(renderer(termenv.ANSI256), width, liveLinks)
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

// The static /cv/ page has to show a terminal nothing but the résumé: all of
// the HTML must sit inside one OSC string, which terminals swallow.
func TestCVPageHidesItsHTMLFromTerminals(t *testing.T) {
	page := cvPage()
	if !strings.HasPrefix(page, "\x1b]9999;<!doctype html>") {
		t.Fatalf("page must open with the OSC and the doctype, got %q", page[:40])
	}
	end := strings.IndexByte(page, '\a')
	if end < 0 {
		t.Fatal("OSC is never terminated")
	}
	head := page[len("\x1b]9999;"):end]
	if strings.ContainsAny(head, "\n\r\x1b") {
		t.Fatal("the HTML head must be one line with no control characters, or a terminal ends the OSC early and prints the rest")
	}
	if len(head) > 1024 {
		t.Fatalf("head is %d bytes; keep the OSC short, some terminals cap it", len(head))
	}
	shown := page[end+1:]
	if strings.Contains(shown, "<") {
		t.Fatal("markup outside the OSC would be printed in the terminal")
	}
	if !strings.Contains(shown, "ssh mosambiswas.com") {
		t.Fatal("the static page should point at the ssh app, now that it is live")
	}
}

// The SSH app is an easter egg too; its claim link has to be reachable
// both inside the app and in the scrollback it leaves behind.
func TestSSHAppHandsOutItsClaimLink(t *testing.T) {
	m := newModel(renderer(termenv.Ascii), func(string) {})
	if !strings.Contains(m.contactPage(80), "/?ssh") {
		t.Fatal("the Contact tab should carry the ?ssh claim link")
	}
}

// Standing rule for everything Mosam publishes: no em or en dashes.
func TestNoDashesInCopy(t *testing.T) {
	text := resumeText(renderer(termenv.Ascii), 78, liveLinks)
	m := newModel(renderer(termenv.Ascii), func(string) {})
	text += m.aboutPage(80) + m.contactPage(80)
	for _, p := range portfolio {
		text += m.detail(p, 60)
	}
	for _, w := range petMoods {
		text += petCaption(w)
	}
	text += petText(time.Now()) + m.switchTo(tabPet).petPage(80, 30)
	if i := strings.IndexAny(text, "—–"); i >= 0 {
		t.Fatalf("dash found near %q", text[max(i-30, 0):min(i+30, len(text))])
	}
}

// ---------- the pet ----------

// Every mood the bridge can report, and the stale one.
var petMoods = []petWorld{
	{}, {act: "code"}, {act: "code", media: "music"}, {act: "game"},
	{act: "game", media: "music"}, {media: "music"}, {media: "watch"}, {away: true},
}

func postPet(t *testing.T, h http.Handler, auth, body string, overTLS bool) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/pet", strings.NewReader(body))
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	if overTLS {
		req.TLS = &tls.ConnectionState{}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code
}

// The pet tab is public, so the only way in is the bridge's key, over
// https, carrying one of the moods and nothing else.
func TestPetTakesAMoodOnlyFromItsKey(t *testing.T) {
	petLatest.Store(nil)
	t.Cleanup(func() { petLatest.Store(nil) })
	sum := sha256.Sum256([]byte("k3y"))
	h := withPetBeacon(webHandler("https://www.mosambiswas.com"), sum[:], true)

	for _, c := range []struct {
		auth, body string
		tls        bool
		want       int
	}{
		{"", "code -", true, http.StatusUnauthorized},
		{"Bearer nope", "code -", true, http.StatusUnauthorized},
		{"k3y", "code -", true, http.StatusUnauthorized}, // no Bearer
		{"Bearer k3y", "code -", false, http.StatusForbidden},
		{"Bearer k3y", "Visual Studio Code - secret.py", true, http.StatusBadRequest},
		{"Bearer k3y", "code music now", true, http.StatusBadRequest},
		{"Bearer k3y", strings.Repeat("a", 65), true, http.StatusRequestEntityTooLarge},
	} {
		if got := postPet(t, h, c.auth, c.body, c.tls); got != c.want {
			t.Fatalf("%q %q tls=%v: got %d, want %d", c.auth, c.body, c.tls, got, c.want)
		}
	}
	if petLatest.Load() != nil {
		t.Fatal("a refused report must change nothing")
	}

	if got := postPet(t, h, "Bearer k3y", "code music", true); got != http.StatusNoContent {
		t.Fatalf("the right key: got %d", got)
	}
	if w := petMood(time.Now()); w != (petWorld{act: "code", media: "music"}) {
		t.Fatalf("mood after the report: %+v", w)
	}
}

// Twenty minutes with nothing from the PC is a PC that is off: the pet
// must not go on claiming he is coding.
func TestPetGoesAwayWhenThePCGoesQuiet(t *testing.T) {
	t.Cleanup(func() { petLatest.Store(nil) })
	now := time.Now()
	petLatest.Store(nil)
	if !petMood(now).away {
		t.Fatal("no report yet should be away")
	}
	petLatest.Store(&petReport{world: petWorld{act: "game"}, at: now.Add(-5 * time.Minute)})
	if petMood(now).act != "game" {
		t.Fatal("a five minute old report still holds")
	}
	if !petMood(now.Add(16 * time.Minute)).away {
		t.Fatal("past twenty minutes it should be away")
	}
}

// A hash that does not parse turns the reports off rather than open, and
// the one compiled in must parse, or the live pet never hears the PC.
func TestPetKeyHash(t *testing.T) {
	if petKeyHash(petKeySHA256) == nil {
		t.Fatal("petKeySHA256 is not a SHA-256 in hex")
	}
	for _, bad := range []string{"", "zz", hex.EncodeToString([]byte("short"))} {
		if petKeyHash(bad) != nil {
			t.Fatalf("%q should turn the reports off", bad)
		}
	}
	h := withPetBeacon(webHandler("https://www.mosambiswas.com"), nil, true)
	if got := postPet(t, h, "Bearer anything", "code -", true); got != http.StatusNotFound {
		t.Fatalf("with no hash: got %d, want 404", got)
	}
}

// curl gets one frame that fits the terminal; a browser gets the handbook.
func TestCurlGetsThePet(t *testing.T) {
	rec := get(t, "/pet", "curl/8.9.1")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "ssh mosambiswas.com") {
		t.Fatalf("got %d %q", rec.Code, rec.Body.String())
	}
	if !strings.ContainsRune(rec.Body.String(), 0x28ff) && !strings.ContainsAny(rec.Body.String(), "⣿⣶⣤") {
		t.Fatal("no face in the body")
	}
	for i, line := range strings.Split(rec.Body.String(), "\n") {
		if w := ansi.StringWidth(line); w > 78 {
			t.Fatalf("line %d is %d wide", i+1, w)
		}
	}
	agent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/140.0"
	rec = get(t, "/pet", agent)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "https://www.mosambiswas.com/Desktop-pet/" {
		t.Fatalf("browser: got %d %q", rec.Code, rec.Header().Get("Location"))
	}
}

// A 24 row terminal leaves 17 rows for the page: the whole face and the
// line that says what he is doing must both be there.
func TestPetTabFitsASmallTerminal(t *testing.T) {
	petLatest.Store(&petReport{world: petWorld{act: "code"}, at: time.Now()})
	t.Cleanup(func() { petLatest.Store(nil) })
	for _, size := range [][2]int{{80, 24}, {minWidth, 24}, {100, 40}} {
		m := newModel(renderer(termenv.Ascii), func(string) {})
		m.width, m.height = size[0], size[1]
		m = m.switchTo(tabPet)
		m.pet.advance(100)
		view := m.View()
		if !strings.Contains(view, "Writing code.") {
			t.Fatalf("%dx%d: the caption is missing", size[0], size[1])
		}
		if n := strings.Count(view, "\n") + 1; n > size[1] {
			t.Fatalf("%dx%d: %d lines", size[0], size[1], n)
		}
	}
}

// Five minutes of every mood, petted now and then: nothing the firmware's
// shapes do may ever fall off the screen and panic.
func TestPetRunsEveryMood(t *testing.T) {
	for _, w := range petMoods {
		p := newPetFace(7, time.Now)
		p.world = w
		for i := 0; i < 15000; i++ {
			if i%700 == 0 {
				p.touch()
			}
			p.advance(1)
		}
	}
}
