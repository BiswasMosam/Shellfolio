package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// The SSH app: four tabs across the top, the work list as a menu, and the
// rest as scrollable pages. Keyboard only, so visitors can still select text
// with the mouse to copy it.

var tabs = []string{"Work", "About", "Résumé", "Contact"}

const (
	tabWork = iota
	tabAbout
	tabResume
	tabContact
)

const (
	maxWidth  = 100 // past this the layout stops growing and sits left
	splitMin  = 76  // below this the work menu and its detail stack instead
	listWidth = 30
	minWidth  = 44 // the tab bar needs this much
)

var ist = time.FixedZone("IST", 5*3600+30*60)

type (
	tickMsg      time.Time
	toastDoneMsg int
)

type model struct {
	r    *lipgloss.Renderer
	p    palette
	copy func(string) // puts text on the visitor's clipboard, when their terminal allows it

	width, height int
	tab           int
	cursor        int
	open          bool // narrow terminals: showing one project full width
	vp            viewport.Model
	now           time.Time

	toast   string
	toastID int
}

func newModel(r *lipgloss.Renderer, copy func(string)) model {
	return model{r: r, p: newPalette(r), copy: copy, width: 80, height: 24, now: time.Now()}
}

func (m model) Init() tea.Cmd { return tick() }

func tick() tea.Cmd {
	return tea.Tick(15*time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.fill(false)
		return m, nil

	case tickMsg:
		m.now = time.Time(msg)
		if m.tab == tabContact {
			m.fill(false) // the contact page says what time it is here
		}
		return m, tick()

	case toastDoneMsg:
		if int(msg) == m.toastID {
			m.toast = ""
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc", "backspace":
			if m.open {
				m.open = false
				return m, nil
			}
			if msg.String() == "esc" {
				return m, tea.Quit
			}
		case "right", "l", "tab":
			return m.switchTo((m.tab + 1) % len(tabs)), nil
		case "left", "h", "shift+tab":
			return m.switchTo((m.tab + len(tabs) - 1) % len(tabs)), nil
		case "1", "2", "3", "4":
			return m.switchTo(int(msg.String()[0] - '1')), nil
		case "c", "y":
			return m.copyCurrent()
		}

		if m.tab == tabWork {
			switch msg.String() {
			case "up", "k":
				m.cursor = (m.cursor + len(portfolio) - 1) % len(portfolio)
			case "down", "j":
				m.cursor = (m.cursor + 1) % len(portfolio)
			case "home", "g":
				m.cursor = 0
			case "end", "G":
				m.cursor = len(portfolio) - 1
			case "enter", " ":
				m.open = m.layoutWidth() < splitMin
			}
			return m, nil
		}

		switch msg.String() {
		case "home", "g":
			m.vp.GotoTop()
			return m, nil
		case "end", "G":
			m.vp.GotoBottom()
			return m, nil
		}
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) switchTo(t int) model {
	m.tab, m.open = t, false
	m.fill(true)
	return m
}

func (m model) copyCurrent() (tea.Model, tea.Cmd) {
	var text, what string
	switch m.tab {
	case tabWork:
		text, what = portfolio[m.cursor].URL, "link"
		if text == "" {
			return m, nil
		}
	case tabContact:
		text, what = email, "email"
	default:
		return m, nil
	}
	m.copy(text)
	m.toastID++
	m.toast = what + " copied, if your terminal allows it"
	id := m.toastID
	return m, tea.Tick(2500*time.Millisecond, func(time.Time) tea.Msg { return toastDoneMsg(id) })
}

// fill sizes the scrolling page and loads the text for the current tab.
func (m *model) fill(top bool) {
	w, h := m.layoutWidth(), m.bodyHeight()
	offset := m.vp.YOffset
	m.vp = viewport.New(w, h)
	switch m.tab {
	case tabAbout:
		m.vp.SetContent(m.aboutPage(w))
	case tabResume:
		m.vp.SetContent(resumeText(m.r, w, nil))
	case tabContact:
		m.vp.SetContent(m.contactPage(w))
	}
	if !top {
		m.vp.SetYOffset(offset)
	}
}

func (m model) layoutWidth() int { return min(m.width, maxWidth) }

// header 2 lines, gap 1, tabs 2, footer 2
func (m model) bodyHeight() int { return max(m.height-7, 3) }

func (m model) View() string {
	w := m.layoutWidth()
	if w < minWidth || m.height < 12 {
		return "\n  " + m.p.accent.Render("Mosam Biswas") + "\n\n  " +
			m.p.dim.Render("Make the window a little bigger,") + "\n  " +
			m.p.dim.Render("or press q to leave.")
	}

	var body string
	switch {
	case m.tab == tabWork && w >= splitMin:
		body = m.workSplit(w)
	case m.tab == tabWork && m.open:
		body = cut(m.detail(portfolio[m.cursor], w-4), m.bodyHeight(), "  ")
	case m.tab == tabWork:
		body = m.workList(w)
	default:
		body = m.vp.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.header(w),
		"",
		m.tabBar(w),
		fit(body, m.bodyHeight()),
		m.footer(w),
	)
}

func (m model) header(w int) string {
	left := "  " + m.p.bold.Render(strings.ToUpper(name))
	right := m.p.dim.Render("Navi Mumbai · "+m.now.In(ist).Format("15:04")+" IST") + "  "
	line1 := spread(left, right, w)
	return line1 + "\n" + ansi.Truncate("  "+m.p.dim.Render(role), w, "")
}

func (m model) tabBar(w int) string {
	var labels, under strings.Builder
	labels.WriteString("  ")
	under.WriteString(m.p.dim.Render("──"))
	for i, t := range tabs {
		label := fmt.Sprintf("%d %s", i+1, t)
		if i == m.tab {
			labels.WriteString(m.p.accent.Bold(true).Render(label))
			under.WriteString(m.p.accent.Render(strings.Repeat("━", ansi.StringWidth(label))))
		} else {
			labels.WriteString(m.p.dim.Render(label))
			under.WriteString(m.p.dim.Render(strings.Repeat("─", ansi.StringWidth(label))))
		}
		if i < len(tabs)-1 {
			labels.WriteString("   ")
			under.WriteString(m.p.dim.Render("───"))
		}
	}
	rest := w - 2 - ansi.StringWidth(under.String())
	if rest > 0 {
		under.WriteString(m.p.dim.Render(strings.Repeat("─", rest)))
	}
	return labels.String() + "\n" + under.String()
}

func (m model) footer(w int) string {
	keys := "←→ section  "
	switch {
	case m.tab == tabWork && m.open:
		keys += "esc back  c copy link"
	case m.tab == tabWork && w < splitMin:
		keys += "↑↓ choose  enter open  c copy link"
	case m.tab == tabWork:
		keys += "↑↓ choose  c copy link"
	case m.tab == tabContact:
		keys += "↑↓ scroll  c copy email"
	default:
		keys += "↑↓ scroll"
	}
	keys += "  q quit"

	right := ""
	if m.toast != "" {
		right = m.p.accent.Render(m.toast) + "  "
	} else if m.tab != tabWork && m.vp.TotalLineCount() > m.vp.Height {
		right = m.p.dim.Render(fmt.Sprintf("%3.f%%", m.vp.ScrollPercent()*100)) + "  "
	}
	if ansi.StringWidth(keys)+ansi.StringWidth(right)+4 > w {
		keys = "←→ ↑↓ c q"
	}
	return "\n" + spread("  "+m.p.dim.Render(keys), right, w)
}

// ---------- work ----------

func (m model) workList(w int) string {
	h := m.bodyHeight()
	lines := make([]string, 0, len(portfolio))
	for i, pr := range portfolio {
		lines = append(lines, m.listRow(i, pr, w))
	}
	// Keep the chosen row on screen when the list is taller than the room.
	start := 0
	if m.cursor >= h {
		start = m.cursor - h + 1
	}
	end := min(start+h, len(lines))
	return strings.Join(lines[start:end], "\n")
}

func (m model) listRow(i int, pr project, w int) string {
	idx := fmt.Sprintf("%02d", i+1)
	title := pr.Name
	badge := ""
	if pr.Badge != "" {
		badge = " " + m.p.accent.Render(strings.ToUpper(pr.Badge))
	}
	if i == m.cursor {
		return m.p.accent.Render("› "+idx) + "  " + m.p.bold.Render(title) + badge
	}
	return "  " + m.p.dim.Render(idx) + "  " + title + badge
}

func (m model) workSplit(w int) string {
	h := m.bodyHeight()
	left := strings.Split(m.workList(listWidth), "\n")
	right := strings.Split(cut(m.detail(portfolio[m.cursor], w-listWidth-5), h, ""), "\n")
	var b strings.Builder
	for i := 0; i < h; i++ {
		l, r := "", ""
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		b.WriteString(pad(l, listWidth) + m.p.dim.Render("│") + "  " + r)
		if i < h-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// detail is one project, width columns wide, with no left margin of its own.
func (m model) detail(pr project, width int) string {
	var b strings.Builder
	b.WriteString(spread(m.p.bold.Render(strings.ToUpper(pr.Name)), m.p.dim.Render(pr.Year), width) + "\n")
	para(&b, "", width, pr.Tagline, m.p.dim.Render)
	b.WriteString("\n")
	b.WriteString(m.p.accent.Bold(true).Render(pr.Stat) + "\n")
	para(&b, "", width, pr.StatFor, m.p.dim.Render)
	b.WriteString("\n")
	para(&b, "", width, pr.About, plain)
	b.WriteString("\n")
	labelled(&b, "", 7, width, m.p.dim.Render("Stack"), pr.Stack, plain)
	if pr.URL != "" {
		labelled(&b, "", 7, width, m.p.dim.Render("Link"), pr.URL, m.p.accent.Render)
	}
	return strings.TrimRight(b.String(), "\n")
}

// ---------- pages ----------

func (m model) aboutPage(w int) string {
	width := min(w, 76)
	var b strings.Builder
	b.WriteString("\n")
	for _, p := range aboutText {
		para(&b, "  ", width, p, plain)
		b.WriteString("\n")
	}
	for _, n := range numbers {
		labelled(&b, "  ", 6, width, m.p.accent.Bold(true).Render(n.N), n.What, m.p.dim.Render)
	}

	b.WriteString("\n  " + m.p.accent.Render("Published at IEEE") + "\n\n")
	para(&b, "  ", width, research.Title, m.p.bold.Render)
	para(&b, "  ", width, research.Role+" · "+research.Venue, m.p.dim.Render)
	b.WriteString("\n")
	para(&b, "  ", width, research.Summary, plain)
	para(&b, "  ", width, research.URL, m.p.accent.Render)
	return b.String()
}

func (m model) contactPage(w int) string {
	var b strings.Builder
	b.WriteString("\n  " + m.p.bold.Render("Say hello.") + "\n\n")
	rows := [][2]string{
		{"Email", email},
		{"Site", site},
		{"GitHub", github},
		{"LinkedIn", linkedin},
		{"Résumé", pdf},
	}
	for _, r := range rows {
		labelled(&b, "  ", 10, w, m.p.dim.Render(r[0]), r[1], m.p.accent.Render)
	}
	b.WriteString("\n")
	para(&b, "  ", w, status+".", plain)
	para(&b, "  ", w, "Based in "+place+". It is "+m.now.In(ist).Format("3:04 pm")+" here.", m.p.dim.Render)

	b.WriteString("\n  " + m.p.accent.Render("Other ways in") + "\n\n")
	ways := [][2]string{
		{"curl mosambiswas.com", "this résumé, as text"},
		{"www.mosambiswas.com/humans.txt", "who and what made the site"},
	}
	for _, r := range ways {
		labelled(&b, "  ", 34, w, r[0], r[1], m.p.dim.Render)
	}

	b.WriteString("\n  " + m.p.accent.Render("You found an easter egg") + "\n\n")
	para(&b, "  ", w, "This app is one of the site's easter eggs. Open this link in a browser to claim it:", m.p.dim.Render)
	para(&b, "  ", w, sshClaim, m.p.accent.Render)
	return b.String()
}

// labelled writes a label column then a value that wraps inside its own
// column. When there is no room for a column, the value goes underneath.
func labelled(b *strings.Builder, indent string, col, width int, label, value string, st style) {
	room := width - len(indent) - col
	if room < 16 || ansi.StringWidth(label) >= col {
		b.WriteString(indent + label + "\n")
		para(b, indent, width, value, st)
		return
	}
	for i, l := range wrap(value, room) {
		lead := strings.Repeat(" ", col)
		if i == 0 {
			lead = label + strings.Repeat(" ", col-ansi.StringWidth(label))
		}
		b.WriteString(indent + lead + st(l) + "\n")
	}
}

// ---------- layout helpers ----------

// spread puts left and right on one line of width columns.
func spread(left, right string, width int) string {
	gap := width - ansi.StringWidth(left) - ansi.StringWidth(right)
	return left + strings.Repeat(" ", max(gap, 1)) + right
}

// cut keeps the first n lines of s, each prefixed with margin.
func cut(s string, n int, margin string) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	for i := range lines {
		lines[i] = margin + lines[i]
	}
	return strings.Join(lines, "\n")
}

// fit pads or trims s to exactly n lines so the footer never moves.
func fit(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	for len(lines) < n {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
