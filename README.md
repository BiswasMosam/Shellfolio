<div align="center">

# Shellfolio

**My portfolio, for people who live in a terminal.**

[![Tests](https://github.com/BiswasMosam/Shellfolio/actions/workflows/tests.yml/badge.svg)](https://github.com/BiswasMosam/Shellfolio/actions/workflows/tests.yml)

```
ssh mosambiswas.com     an interactive menu of everything I've built
curl mosambiswas.com    the résumé, as text
```

One small Go binary answers both, **live since 27 September 2026**. Browsers
that land on the bare domain are sent on to [www.mosambiswas.com](https://www.mosambiswas.com),
where the real site lives.

</div>

---

## What you get

### `curl mosambiswas.com`

The one page résumé, set for an 80 column terminal, in the site's vermilion:

```
  ███╗   ███╗ ██████╗ ███████╗ █████╗ ███╗   ███╗
  ████╗ ████║██╔═══██╗██╔════╝██╔══██╗████╗ ████║
  ██╔████╔██║██║   ██║███████╗███████║██╔████╔██║
  ██║╚██╔╝██║██║   ██║╚════██║██╔══██║██║╚██╔╝██║
  ██║ ╚═╝ ██║╚██████╔╝███████║██║  ██║██║ ╚═╝ ██║
  ╚═╝     ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝
  ██████╗ ██╗███████╗██╗    ██╗ █████╗ ███████╗
  ██╔══██╗██║██╔════╝██║    ██║██╔══██╗██╔════╝
  ██████╔╝██║███████╗██║ █╗ ██║███████║███████╗
  ██╔══██╗██║╚════██║██║███╗██║██╔══██║╚════██║
  ██████╔╝██║███████║╚███╔███╔╝██║  ██║███████║
  ╚═════╝ ╚═╝╚══════╝ ╚══╝╚══╝ ╚═╝  ╚═╝╚══════╝

  Software engineer · AI / ML · Full-stack
  B.Tech AI & DS · Class of 2026 · Navi Mumbai, India
  ● Open to full-time roles & internships

  (01) EXPERIENCE ────────────────────────────────────────────────────────────

  Software Developer Intern                                     Jan / Apr 2026
  Sedna Technologies · Mumbai
  · Shipped work across 4 client projects, working directly with clients to
    gather requirements and resolve issues.
  ...
```

The name is stacked the way the website's hero sets it, solid blocks in the
terminal's own colour and the box-drawn shadow in vermilion, so the letters sit
on a thin line of embers.

| Address | Returns |
|---|---|
| `curl mosambiswas.com` | the résumé with colour (256 colour escapes, which every modern terminal reads) |
| `curl mosambiswas.com/plain` | the same with no escapes, for saving to a file |
| `curl mosambiswas.com/cv` | the same as the first, kept for links from before the server |
| `/llms.txt`, `/llms-full.txt`, `/humans.txt`, `/robots.txt` | a 301 to the same file on www, with a line saying where it is (`curl -L` follows it) |
| `curl mosambiswas.com/pet` | my desk pet, one frame of it, doing what I am doing (see the Pet tab below). A browser asking for `/pet` is sent to the pet's handbook |
| anything else, from a terminal | a short 404 that points back to the two above |

`wget`, `httpie` and `xh` get the same treatment as `curl`.

### The static copy on www

Before this server existed, the curl trick ran on GitHub Pages alone, and that
copy is still there at `www.mosambiswas.com/cv/`: one static file,
`/cv/index.html` in the
[site repo](https://github.com/BiswasMosam/BiswasMosam.github.io). That one
file reads right in two places:

```
ESC ] 9999 ; <!doctype html> … <pre id="cv"> BEL   ← terminals swallow this
  ███╗   ███╗ ██████╗ …                            ← and print only this
```

The HTML head sits inside an OSC escape sequence. Terminals drop OSC strings
they don't recognise without printing them, so `curl` shows only the résumé.
A browser reads the same bytes as HTML: the head loads a stylesheet that hides
the raw text and a script that turns the colour codes into styled spans. The
closing tags are left off on purpose; HTML doesn't need them, and a terminal
would print them.

`go run . -cv <path>` writes that file (see [`cvpage.go`](cvpage.go)); regenerate it
whenever the résumé changes. A test checks that all the markup stays inside the
escape sequence, on one line, short enough for terminals that cap OSC length.

### `ssh mosambiswas.com`

A full-screen app with five tabs:

| Tab | What is on it |
|---|---|
| **1 Work** | All eleven projects from the homepage as a menu. The selected one opens beside the list: its headline number, what it is, the stack, the link. Below 76 columns the list and the detail stack instead, and `enter` opens one. |
| **2 About** | The homepage's about section, the numbers, and the IEEE paper. |
| **3 Résumé** | The same text `curl` gets, scrollable. |
| **4 Contact** | Email, links, and what time it is in Navi Mumbai right now. |
| **5 Pet** | My [desk pet](https://www.mosambiswas.com/Desktop-pet/), live, drawn in braille. It codes, games, dances to music or eats popcorn through a film when I do, a few minutes behind, and sleeps when I am away. `space` pets it. |

| Keys | |
|---|---|
| `←` `→` · `tab` · `1` to `5` | switch tabs |
| `↑` `↓` · `j` `k` | choose a project, or scroll |
| `g` `G` · `home` `end` | top, bottom |
| `enter` · `esc` | open a project, back (narrow terminals) |
| `c` | copy the project's link, or my email on Contact, to **your** clipboard |
| `space` | pet the pet (three quick ones and it falls for you) |
| `q` | leave |

Copying works over SSH through OSC 52: the app sends the text inside an escape
sequence and your terminal puts it on your local clipboard. Windows Terminal,
iTerm2, kitty, WezTerm and Alacritty do this; terminals that don't simply
ignore it. tmux and screen are detected and wrapped.

### The pet

The face on the Pet tab is the desk pet's firmware
([Desktop-pet](https://github.com/BiswasMosam/Desktop-pet), `face.cpp` and its
props) ported to Go line by line, down to the Adafruit GFX routines that draw
its rounded eyes, so it glances, blinks, fidgets and eases exactly as the one
on my desk does. Braille characters hold two by four pixels each, so the whole
128 by 64 screen fits in 64 columns by 16 rows with nothing lost.

What it does comes from the PC beside the real pet, which reports one mood:
code or a game in front, music or a film playing, at the desk, or away. Never
a window title, a song or a program's name. A change is sent at most every
five minutes, and twenty minutes of silence reads as away. The report carries
a key that only that PC holds; the server keeps just the key's SHA-256
(`petKeySHA256` in [`pet.go`](pet.go)), so publishing it gives nothing away.

The app is keyboard only on purpose. Capturing the mouse would stop you from
selecting text to copy it.

`ssh mosambiswas.com resume`, or any SSH session without a terminal (a pipe, a
script), gets the plain text résumé instead of the app.

### Easter eggs

Both doors are among the [seventeen easter eggs](https://www.mosambiswas.com/help)
on the site, which a browser keeps count of. A terminal can't reach that count,
so each hands you a link that claims its egg when opened in a browser: the
résumé's last line (`/?curl`), and for the app, the Contact tab and the goodbye
line it leaves in your scrollback (`/?ssh`).

---

## How it works

```
                    ┌──────────────────────────────────────────┐
  mosambiswas.com ──►  Shellfolio (one Go binary on a VPS)     │
     (A record)     │                                          │
                    │  :22   Wish SSH server ─► Bubble Tea app  │
                    │  :80   curl/wget  ─► résumé as text       │
                    │  :443  browsers   ─► 301 to www, path     │
                    │                        and query intact   │
                    └──────────────────────────────────────────┘

  www.mosambiswas.com ──► GitHub Pages (unchanged)
     (CNAME)
```

Before this, the bare domain pointed at GitHub Pages, whose only job there was
to redirect to www. Shellfolio takes that job over and keeps the redirect
identical, so every old link to `mosambiswas.com/...` still lands where it did.
The website itself, and the installable app on www, are not touched.

HTTPS on the bare domain comes from Let's Encrypt through Go's `autocert`,
fetched on the first request and renewed on its own. Port 80 stays plain HTTP
rather than redirecting, because that is what `curl mosambiswas.com` speaks.

### Files

| File | What it does |
|---|---|
| [`content.go`](content.go) | Every word a visitor can read. Two lists on purpose: `resumeWork` mirrors the one page résumé, `portfolio` mirrors the homepage's work list. |
| [`text.go`](text.go) | Renders the résumé as terminal text at any width from 40 to 80 columns. Feeds curl, the Résumé tab, and SSH sessions without a terminal. |
| [`banner.go`](banner.go) | The five letters of the ANSI Shadow figlet face the name needs, and the colouring that splits blocks from shadow. |
| [`tui.go`](tui.go) | The Bubble Tea model for the SSH app. |
| [`pet.go`](pet.go) | The Pet tab, `curl /pet`, and the report the PC sends (`POST /pet`, key checked by hash, one mood and nothing else). |
| [`petface.go`](petface.go) | The desk pet's face and props, ported from its firmware. |
| [`gfx.go`](gfx.go) | A 128x64 one bit screen with the Adafruit GFX shapes the firmware uses, and the braille that draws it. |
| [`web.go`](web.go) | The HTTP handler: terminals get text, everyone else a 301 to www. The site's own text files (`siteFiles`) send terminals to www too. |
| [`cvpage.go`](cvpage.go) | Writes the static `/cv/` page: the résumé for terminals with its HTML hidden in an escape sequence. |
| [`main.go`](main.go) | Starts the three listeners, the SSH middleware, and a clean shutdown. |
| [`deploy/shellfolio.service`](deploy/shellfolio.service) | systemd unit: runs as its own user with only `CAP_NET_BIND_SERVICE`, can write nothing but its own folder. |

### Looking after a public server

Anyone can connect, so the SSH side is built for strangers:

- **No shell, ever.** Every session runs the portfolio app or prints text. There is nothing else to reach.
- **No login.** No passwords or keys are asked for, and none are stored.
- **Rate limited** per IP (one new session every 2 seconds, bursts of 5), **capped** at 64 open sessions, closed after 15 idle minutes and at most an hour.
- **One way in that is not for visitors:** `POST /pet`, which takes a mood from the PC beside the desk pet. It needs a key only that PC holds (compared by hash, in constant time), only over HTTPS, at most 64 bytes, and one of eight fixed moods. Anything else is refused and changes nothing.
- **Unprivileged.** The service holds exactly one capability, binding low ports. The filesystem is read-only to it apart from `/var/lib/shellfolio`.

---

## Run it locally

Needs Go 1.27 or newer (the server itself needs no Go: build locally and copy the binary).

```bash
git clone https://github.com/BiswasMosam/Shellfolio.git
cd Shellfolio
go run .
```

Defaults are development ports, so nothing needs admin rights:

```bash
ssh -p 2222 localhost          # the app
curl localhost:8080            # the résumé
curl localhost:8080/plain
```

The first run writes an SSH host key to `.ssh/id_ed25519`. It is gitignored.

| Flag | Default | |
|---|---|---|
| `-ssh` | `:2222` | SSH address, `:22` in production |
| `-http` | `:8080` | HTTP address, `:80` in production |
| `-https` | off | HTTPS address, e.g. `:443`. Needs `-domain` |
| `-domain` | | domain to fetch a certificate for |
| `-certs` | `certs` | where certificates are kept |
| `-hostkey` | `.ssh/id_ed25519` | the SSH host key. **Keep it.** A new one makes every returning visitor's `ssh` warn about a changed key |
| `-www` | `https://www.mosambiswas.com` | where browsers are sent |
| `-max-sessions` | `64` | SSH sessions open at once |
| `-cv` | | write the static `/cv/` page to this path and exit, instead of serving |
| `-pet-key` | `petKeySHA256` | SHA-256 in hex of the key the pet's PC reports with. Empty turns the reports off |

### Tests

```bash
go test ./...
```

GitHub runs them on every push (`.github/workflows/tests.yml`), along with
`go vet` and the Linux build the server runs. They check that terminals get the résumé and browsers get the exact redirect,
that terminals asking for `llms.txt` and the other text files are sent to www,
that `/plain` carries no escapes, that **every line of every tab fits** every
window from 44 to 140 columns (a line one column too wide wraps in the
visitor's terminal and the layout falls apart), that no em or en dash
appears anywhere in the copy, and that the pet's report door takes only the
right key over HTTPS with one of its moods, goes to sleep when the PC goes
quiet, and runs five minutes of every mood without drawing off the screen.

---

## Deploying

What it needs: any small Linux server with a public IPv4 address. Oracle
Cloud's Always Free tier is enough and costs nothing; any $4 to $6 a month VPS
works the same. Ubuntu is assumed below.

**1. Move your own SSH out of the way first.** Visitors will get port 22, so
the admin login moves to 2222. Open 2222 in the provider's firewall, then:

```bash
# Ubuntu 22.10 and later (ssh is socket activated)
sudo mkdir -p /etc/systemd/system/ssh.socket.d
printf '[Socket]\nListenStream=\nListenStream=2222\n' | sudo tee /etc/systemd/system/ssh.socket.d/port.conf
sudo systemctl daemon-reload && sudo systemctl restart ssh.socket

# older Ubuntu: set "Port 2222" in /etc/ssh/sshd_config, then
sudo systemctl restart ssh
```

In a **second** window, check `ssh -p 2222 you@server` works before closing the
first. Then open 22, 80 and 443 in the provider's firewall. Some images ship
their own iptables rules as well (Oracle Cloud's Ubuntu images do), so allow
80, 443 and 2222 there too.

**2. Build and copy.**

```bash
GOOS=linux GOARCH=amd64 go build -o dist/shellfolio .   # arm64 for Ampere / Graviton
scp -P 2222 dist/shellfolio deploy/shellfolio.service you@server:
```

From PowerShell: `$env:GOOS='linux'; $env:GOARCH='amd64'; go build -o dist/shellfolio .`

**3. Install it as a service.**

```bash
sudo useradd --system --home /var/lib/shellfolio --shell /usr/sbin/nologin shellfolio
sudo mkdir -p /var/lib/shellfolio && sudo chown shellfolio: /var/lib/shellfolio
sudo install -m 755 shellfolio /usr/local/bin/shellfolio
sudo install -m 644 shellfolio.service /etc/systemd/system/
sudo systemctl daemon-reload && sudo systemctl enable --now shellfolio
```

**4. Point the bare domain at the server.** At the registrar (Namecheap,
Advanced DNS): delete the four `@` A records that point at GitHub
(`185.199.108.153` to `185.199.111.153`), and add one `@` A record with the
server's IP. Delete any `@` AAAA records too, or IPv6 visitors will still reach
GitHub. **Leave the `www` CNAME alone.**

**5. Check.**

```bash
curl mosambiswas.com
ssh mosambiswas.com
curl -sI https://mosambiswas.com -A Mozilla   # 301 to https://www.mosambiswas.com/
```

The first HTTPS request takes a few seconds while the certificate is issued.

**Updating** is steps 2 and 3's `install` line, then `sudo systemctl restart shellfolio`.
Logs: `journalctl -u shellfolio -f`.

---

## Keeping it true

`content.go` copies two sources, and each list should change when its source
does:

- `resumeWork`, `experience`, `research`, `education`, `stack`, `leadership`
  follow the one page B/W résumé.
- `portfolio` follows the work list on the homepage, including the headline
  number on each project's preview card.

After any change, regenerate the static page and ship it with the site:

```bash
go run . -cv ../BiswasMosam.github.io/cv/index.html
```

Numbers stay the ones the résumé states. Nothing here claims more than the
site does.

---

## Built with

[Go](https://go.dev) · [Wish](https://github.com/charmbracelet/wish) (SSH apps) ·
[Bubble Tea](https://github.com/charmbracelet/bubbletea) (the TUI) ·
[Lip Gloss](https://github.com/charmbracelet/lipgloss) (styling) ·
[Bubbles](https://github.com/charmbracelet/bubbles) (the scrolling pages) ·
`golang.org/x/crypto/acme/autocert` (certificates) ·
[Adafruit GFX](https://github.com/adafruit/Adafruit-GFX-Library) (the pet's shapes, ported; BSD licence)

<div align="center">

**[mosambiswas.com](https://www.mosambiswas.com)** · **[mosambiswas999@gmail.com](mailto:mosambiswas999@gmail.com)**

</div>
