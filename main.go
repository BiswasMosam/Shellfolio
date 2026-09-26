// Shellfolio serves Mosam Biswas's portfolio to terminals.
//
//	ssh mosambiswas.com    an interactive menu of the work
//	curl mosambiswas.com   the résumé as text
//
// One binary answers both. Browsers that land on the bare domain are sent on
// to www, where the real site lives on GitHub Pages.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aymanbagabas/go-osc52/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/charmbracelet/wish/ratelimiter"
	"github.com/muesli/termenv"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/time/rate"
)

func main() {
	var (
		sshAddr   = flag.String("ssh", ":2222", "SSH listen address (:22 in production)")
		httpAddr  = flag.String("http", ":8080", "HTTP listen address (:80 in production)")
		httpsAddr = flag.String("https", "", "HTTPS listen address, e.g. :443. Needs -domain")
		domain    = flag.String("domain", "", "domain to fetch a Let's Encrypt certificate for, e.g. mosambiswas.com")
		certDir   = flag.String("certs", "certs", "where Let's Encrypt certificates are kept")
		hostKey   = flag.String("hostkey", ".ssh/id_ed25519", "SSH host key, created on first run. Keep it: a new one makes every returning visitor's ssh complain")
		www       = flag.String("www", "https://www.mosambiswas.com", "where browsers are redirected")
		maxConns  = flag.Int("max-sessions", 64, "most SSH sessions open at once")
	)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sshSrv, err := newSSHServer(*sshAddr, *hostKey, *maxConns)
	if err != nil {
		log.Fatal("ssh server", "err", err)
	}

	web := webHandler(*www)
	var servers []*http.Server
	if *httpsAddr != "" {
		if *domain == "" {
			log.Fatal("-https needs -domain")
		}
		certs := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(*domain),
			Cache:      autocert.DirCache(*certDir),
		}
		// Port 80 answers the Let's Encrypt challenge and still serves curl,
		// which speaks plain http unless told otherwise.
		servers = append(servers,
			newHTTPServer(*httpAddr, certs.HTTPHandler(web)),
			newHTTPServer(*httpsAddr, web))
		servers[1].TLSConfig = certs.TLSConfig()
	} else {
		servers = append(servers, newHTTPServer(*httpAddr, web))
	}

	errs := make(chan error, len(servers)+1)
	go func() {
		log.Info("ssh", "addr", *sshAddr)
		if err := sshSrv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			errs <- fmt.Errorf("ssh: %w", err)
		}
	}()
	for _, srv := range servers {
		go func() {
			log.Info("http", "addr", srv.Addr, "tls", srv.TLSConfig != nil)
			var err error
			if srv.TLSConfig != nil {
				err = srv.ListenAndServeTLS("", "")
			} else {
				err = srv.ListenAndServe()
			}
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				errs <- fmt.Errorf("http %s: %w", srv.Addr, err)
			}
		}()
	}

	select {
	case <-ctx.Done():
		log.Info("shutting down")
	case err := <-errs:
		log.Error("server stopped", "err", err)
	}

	done, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = sshSrv.Shutdown(done)
	for _, srv := range servers {
		_ = srv.Shutdown(done)
	}
}

func newHTTPServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func newSSHServer(addr, hostKey string, maxSessions int) (*ssh.Server, error) {
	// Middleware runs last to first: logging sees every connection, the rate
	// limiter and the session cap turn away floods, and only then does a
	// visitor get the app, or the text version if they brought no terminal.
	return wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(hostKey),
		wish.WithIdleTimeout(15*time.Minute),
		wish.WithMaxTimeout(time.Hour),
		wish.WithMiddleware(
			goodbye(),
			bm.Middleware(app),
			textWithoutTerminal(),
			sessionCap(maxSessions),
			ratelimiter.Middleware(ratelimiter.NewRateLimiter(rate.Every(2*time.Second), 5, 4096)),
			logging.StructuredMiddleware(),
		),
	)
}

// app starts the interactive portfolio for one visitor.
func app(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()
	m := newModel(bm.MakeRenderer(s), clipboard(s, pty.Term))
	m.width, m.height = pty.Window.Width, pty.Window.Height
	m.fill(true)
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

// clipboard sets the visitor's clipboard with OSC 52. Terminals that do not
// support it ignore the sequence.
func clipboard(s ssh.Session, term string) func(string) {
	return func(text string) {
		seq := osc52.New(text)
		switch {
		case strings.HasPrefix(term, "tmux"):
			seq = seq.Tmux()
		case strings.HasPrefix(term, "screen"):
			seq = seq.Screen()
		}
		_, _ = seq.WriteTo(s)
	}
}

// textWithoutTerminal answers sessions that brought no terminal, like
// `ssh mosambiswas.com resume` or a pipe, with the plain text résumé.
func textWithoutTerminal() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			_, _, hasPty := s.Pty()
			if hasPty && len(s.Command()) == 0 {
				next(s)
				return
			}
			profile := termenv.Ascii
			if hasPty {
				profile = termenv.ANSI256
			}
			say(s, resumeText(renderer(profile), 78, true))
		}
	}
}

func goodbye() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			p := newPalette(bm.MakeRenderer(s))
			say(s, "\n  Thanks for stopping by. Say hello: "+p.accent.Render(email)+"\n\n")
			next(s)
		}
	}
}

func sessionCap(n int) wish.Middleware {
	slots := make(chan struct{}, n)
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			select {
			case slots <- struct{}{}:
				defer func() { <-slots }()
				next(s)
			default:
				say(s, "\n  Busy right now. Try again in a minute, or: curl mosambiswas.com\n\n")
			}
		}
	}
}

// say writes text to a session. With a terminal attached the ssh library
// turns each \n into \r\n itself (its emulated PTY), so plain \n is right.
func say(s ssh.Session, text string) {
	_, _ = fmt.Fprint(s, text)
}
