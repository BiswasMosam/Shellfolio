package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Colours. The accent is the site's vermilion. Body text is left in the
// terminal's own foreground so the résumé reads on light and dark themes.
const (
	vermilion = "#ff5227"
	grey      = "245"
)

type style = func(...string) string

type palette struct {
	accent, dim, bold, name lipgloss.Style
}

func newPalette(r *lipgloss.Renderer) palette {
	return palette{
		accent: r.NewStyle().Foreground(lipgloss.Color(vermilion)),
		dim:    r.NewStyle().Foreground(lipgloss.Color(grey)),
		bold:   r.NewStyle().Bold(true),
		name:   r.NewStyle().Bold(true),
	}
}

func plain(s ...string) string { return strings.Join(s, " ") }

// resumeText renders the one page résumé as terminal text, at most width
// columns wide. The same function feeds curl, the SSH résumé tab and SSH
// sessions that have no terminal attached.
func resumeText(r *lipgloss.Renderer, width int, invite bool) string {
	width = clamp(width, 40, 80)
	p := newPalette(r)
	var b strings.Builder
	const in = "  "
	inner := width - len(in)
	text := func(indent, s string, st style) { para(&b, indent, width, s, st) }

	b.WriteString("\n" + banner(p, in, width) + "\n")
	text(in, role, p.bold.Render)
	text(in, degree+" · "+place, p.dim.Render)
	marked(&b, in, inner, p.accent.Render("●"), status)
	b.WriteString("\n")
	text(in, summary, plain)
	b.WriteString("\n")
	for _, l := range []string{email, site, github, linkedin} {
		text(in, l, p.accent.Render)
	}

	section(&b, p, width, "01", "Experience")
	for _, j := range experience {
		row(&b, in, width, p.bold.Render(j.Role), j.When, p.dim.Render)
		text(in, j.Company+" · "+j.Where, p.dim.Render)
		for _, pt := range j.Points {
			marked(&b, in, inner, p.accent.Render("·"), pt)
		}
	}

	section(&b, p, width, "02", "Selected work")
	for i, pr := range resumeWork {
		if i > 0 {
			b.WriteString("\n")
		}
		idx := p.accent.Render(fmt.Sprintf("%02d", i+1))
		row(&b, in, width, idx+"  "+p.bold.Render(pr.Name), pr.Year, p.dim.Render)
		text(in+"    ", pr.About, plain)
		text(in+"    ", pr.Stack, p.dim.Render)
		if pr.URL != "" {
			text(in+"    ", pr.URL, p.accent.Render)
		}
	}
	b.WriteString("\n")
	text(in, "Six of twenty-plus. The rest: "+site+"/#projects", p.dim.Render)

	section(&b, p, width, "03", "Research")
	text(in, research.Title, p.bold.Render)
	text(in, research.Role+" · "+research.Venue, p.dim.Render)
	b.WriteString("\n")
	text(in, research.Summary, plain)
	text(in, research.URL, p.accent.Render)

	section(&b, p, width, "04", "Education")
	row(&b, in, width, p.bold.Render(education.Name), education.When, p.dim.Render)
	text(in, education.Detail, p.dim.Render)

	section(&b, p, width, "05", "Stack")
	const label = 15
	for _, s := range stack {
		for i, l := range wrap(s.Items, inner-label) {
			head := strings.Repeat(" ", label)
			if i == 0 {
				head = p.dim.Render(pad(s.Group, label))
			}
			b.WriteString(in + head + l + "\n")
		}
	}

	section(&b, p, width, "06", "Leadership & community")
	for _, l := range leadership {
		marked(&b, in, inner, p.accent.Render("·"), l)
	}

	b.WriteString("\n" + in + p.dim.Render(strings.Repeat("─", inner)) + "\n")
	pairs := [][2]string{{"PDF", pdf}}
	if invite {
		pairs = append(pairs,
			[2]string{"The interactive version", "ssh mosambiswas.com"},
			[2]string{"Without colours", "curl mosambiswas.com/plain"})
	}
	for _, pr := range pairs {
		row(&b, in, width, p.dim.Render(pr[0]), pr[1], p.accent.Render)
	}
	b.WriteString("\n")
	return b.String()
}

// banner is the stacked figlet name, or a plain bold line when the terminal
// is too narrow to hold it.
func banner(p palette, in string, width int) string {
	lines := append(figlet("MOSAM"), figlet("BISWAS")...)
	widest := 0
	for _, l := range lines {
		widest = max(widest, ansi.StringWidth(l))
	}
	if widest+len(in) > width {
		return in + p.name.Render(strings.ToUpper(name)) + "\n"
	}
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(in + forge(l, p.name.Render, p.accent.Render) + "\n")
	}
	return b.String()
}

// section draws "(01) EXPERIENCE ────" across the width.
func section(b *strings.Builder, p palette, width int, n, title string) {
	num := "(" + n + ")"
	head := " " + strings.ToUpper(title) + " "
	rule := max(width-2-ansi.StringWidth(num+head), 3)
	b.WriteString("\n\n  " + p.accent.Render(num) + p.bold.Render(head) + p.dim.Render(strings.Repeat("─", rule)) + "\n\n")
}

// para wraps text to the room left after indent and styles each line.
func para(b *strings.Builder, indent string, width int, text string, st style) {
	for _, l := range wrap(text, width-len(indent)) {
		b.WriteString(indent + st(l) + "\n")
	}
}

// marked is a hanging paragraph: mark on the first line, the rest indented
// to line up under the text.
func marked(b *strings.Builder, in string, inner int, mark, text string) {
	for i, l := range wrap(text, inner-2) {
		lead := "  "
		if i == 0 {
			lead = mark + " "
		}
		b.WriteString(in + lead + l + "\n")
	}
}

// row puts left and right on one line, right flush with the edge. Left comes
// styled and is short; right comes plain with its style, so when the two do
// not fit side by side right drops to its own lines and wraps.
func row(b *strings.Builder, in string, width int, left, right string, st style) {
	gap := width - len(in) - ansi.StringWidth(left) - ansi.StringWidth(right)
	if gap < 2 {
		b.WriteString(in + left + "\n")
		para(b, in, width, right, st)
		return
	}
	b.WriteString(in + left + strings.Repeat(" ", gap) + st(right) + "\n")
}

// wrap breaks plain text into lines no wider than width, on spaces. A word
// longer than a whole line (a URL on a narrow screen) is split where it must.
func wrap(text string, width int) []string {
	var lines []string
	var line strings.Builder
	for _, word := range strings.Fields(text) {
		for ansi.StringWidth(word) > width {
			if line.Len() > 0 {
				lines = append(lines, line.String())
				line.Reset()
			}
			head := ansi.Truncate(word, width, "")
			lines = append(lines, head)
			word = word[len(head):]
		}
		if word == "" {
			continue
		}
		if line.Len() > 0 && ansi.StringWidth(line.String())+1+ansi.StringWidth(word) > width {
			lines = append(lines, line.String())
			line.Reset()
		}
		if line.Len() > 0 {
			line.WriteByte(' ')
		}
		line.WriteString(word)
	}
	if line.Len() > 0 {
		lines = append(lines, line.String())
	}
	return lines
}

func pad(s string, width int) string {
	return s + strings.Repeat(" ", max(width-ansi.StringWidth(s), 0))
}

func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}
