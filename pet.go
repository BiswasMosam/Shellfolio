package main

// The pet tab: Mosam's desk pet, live. The PC beside the real one (its
// bridge, in AMINAL's repo) reports a mood here every few minutes, and every
// visitor's copy of the face (petface.go) acts it out.
//
// Only a mood ever arrives: code or a game in front, music or a film
// playing, at the desk, or away. Never a window title, a song, a time of
// its own. The PC holds a key; this server knows only the key's SHA-256,
// so the public source and the binary give nothing away.

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// SHA-256 of the key the pet's bridge sends, made on Mosam's PC with
// `python3.12 pet_bridge.py --online-key` on 23 Nov 2026. The key is 32
// random bytes and lives only in Windows' Credential Manager there; its
// hash is safe to publish, since nothing can be worked back from it.
const petKeySHA256 = "526d937972a26e66e65900aa77e960d6f4dfea1ade48d2a8af623fedc8070170"

const (
	// The bridge says its mood again every ten minutes. Twice that with
	// nothing heard means the PC is off or asleep, and so is the pet.
	petStale = 20 * time.Minute

	petTickEvery = 100 * time.Millisecond
	petFrames    = int(petTickEvery / (frameMs * time.Millisecond))
)

type petReport struct {
	world petWorld
	at    time.Time
}

var petLatest atomic.Pointer[petReport]

// petMood is what the pet should be doing at now: the last report, or away
// once that is stale or there never was one.
func petMood(now time.Time) petWorld {
	r := petLatest.Load()
	if r == nil || now.Sub(r.at) > petStale {
		return petWorld{away: true}
	}
	return r.world
}

// parsePetReport reads the bridge's one line: "away", or what is in front
// and what is playing, "-" for nothing ("code music", "- watch", "- -").
// Anything else is refused whole.
func parsePetReport(line string) (petWorld, bool) {
	f := strings.Fields(line)
	if len(f) == 1 && f[0] == "away" {
		return petWorld{away: true}, true
	}
	if len(f) != 2 {
		return petWorld{}, false
	}
	act, media := f[0], f[1]
	switch act {
	case "code", "game":
	case "-":
		act = ""
	default:
		return petWorld{}, false
	}
	switch media {
	case "music", "watch":
	case "-":
		media = ""
	default:
		return petWorld{}, false
	}
	return petWorld{act: act, media: media}, true
}

// petKeyHash turns the -pet-key flag into bytes. Empty or malformed turns
// the beacon off, so a typo can never leave it open to any key.
func petKeyHash(hexSum string) []byte {
	sum, err := hex.DecodeString(strings.TrimSpace(hexSum))
	if err != nil || len(sum) != sha256.Size {
		return nil
	}
	return sum
}

// withPetBeacon takes the bridge's reports (POST /pet) ahead of the rest
// of the site. needTLS refuses them over plain http, where the key would
// cross the internet in the clear.
func withPetBeacon(next http.Handler, keyHash []byte, needTLS bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pet" || r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if keyHash == nil {
			http.Error(w, "not here", http.StatusNotFound)
			return
		}
		if needTLS && r.TLS == nil {
			http.Error(w, "https only", http.StatusForbidden)
			return
		}
		key, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		sum := sha256.Sum256([]byte(key))
		if !ok || subtle.ConstantTimeCompare(sum[:], keyHash) != 1 {
			http.Error(w, "not yours", http.StatusUnauthorized)
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64))
		if err != nil {
			http.Error(w, "too long", http.StatusRequestEntityTooLarge)
			return
		}
		world, ok := parsePetReport(string(body))
		if !ok {
			http.Error(w, "not a mood", http.StatusBadRequest)
			return
		}
		petLatest.Store(&petReport{world: world, at: time.Now()})
		w.WriteHeader(http.StatusNoContent)
	})
}

// ---------- words ----------

func petCaption(w petWorld) string {
	withMusic := ""
	if w.media == "music" {
		withMusic = ", with music on"
	}
	switch {
	case w.away:
		return "Away from the desk."
	case w.act == "code":
		return "Writing code" + withMusic + "."
	case w.act == "game":
		return "Playing a game" + withMusic + "."
	case w.media == "music":
		return "Listening to music."
	case w.media == "watch":
		return "Watching something."
	}
	return "At the desk."
}

const (
	petAbout    = "My desk pet, live. It does what I am doing at my PC, a few minutes behind."
	petHandbook = "www.mosambiswas.com/Desktop-pet"
)

// One frame costs about a millisecond to make (measured: 0.87 ms asleep,
// the longest), and the web side has no rate limit of its own, so a frame
// is kept ten seconds per mood rather than made for every request.
var petTextCache struct {
	sync.Mutex
	world petWorld
	at    time.Time
	text  string
}

// petText is GET /pet for curl: one frame, a few seconds into a visit.
func petText(now time.Time) string {
	world := petMood(now)
	c := &petTextCache
	c.Lock()
	defer c.Unlock()
	if c.text != "" && c.world == world && now.Sub(c.at) < 10*time.Second {
		return c.text
	}
	c.world, c.at, c.text = world, now, petFrameText(world, now)
	return c.text
}

func petFrameText(world petWorld, now time.Time) string {
	p := newPetFace(uint64(now.UnixNano()), time.Now)
	p.world = world
	frames := 150 // three seconds in: eyes open, a glance or two
	if world.away {
		frames = (nightSleepMs + 2000) / frameMs // asleep, as the words say
	}
	p.advance(frames)
	var b strings.Builder
	b.WriteString("\n")
	for _, l := range p.screen.braille() {
		b.WriteString(strings.TrimRight("  "+l, " ") + "\n")
	}
	b.WriteString("\n  " + petCaption(world) + "\n")
	para(&b, "  ", 78, petAbout+" Watch it move: ssh mosambiswas.com, then 5.", plain)
	b.WriteString("\n")
	return b.String()
}

// ---------- the tab ----------

type petTickMsg int

func petTick(gen int) tea.Cmd {
	return tea.Tick(petTickEvery, func(time.Time) tea.Msg { return petTickMsg(gen) })
}

// petPage is the face over its caption. The full screen needs 66 columns
// and 17 rows; anything smaller gets the half size face.
func (m model) petPage(w, h int) string {
	world := petMood(time.Now())
	face := m.pet.screen.braille()
	if w < 2+screenW/2 || h < len(face)+1 {
		face = m.pet.screen.brailleSmall()
	}
	// The face and the caption always; the blank lines around them only
	// when there is room, so a 24 row terminal still gets the caption.
	var b strings.Builder
	if h >= len(face)+3 {
		b.WriteString("\n")
	}
	for _, l := range face {
		b.WriteString("  " + l + "\n")
	}
	if h >= len(face)+2 {
		b.WriteString("\n")
	}
	b.WriteString("  " + m.p.accent.Bold(true).Render(petCaption(world)) + "\n")
	para(&b, "  ", min(w, 76), petAbout, m.p.dim.Render)
	b.WriteString("\n")
	labelled(&b, "  ", 10, w, m.p.dim.Render("The pet"), petHandbook, m.p.accent.Render)
	return strings.TrimRight(b.String(), "\n")
}
