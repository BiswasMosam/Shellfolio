package main

import "strings"

// Just enough of the "ANSI Shadow" figlet face to spell the name. Stacked
// MOSAM over BISWAS, the way the site's hero sets it.
var shadow = map[rune][6]string{
	'M': {
		"███╗   ███╗",
		"████╗ ████║",
		"██╔████╔██║",
		"██║╚██╔╝██║",
		"██║ ╚═╝ ██║",
		"╚═╝     ╚═╝",
	},
	'O': {
		" ██████╗ ",
		"██╔═══██╗",
		"██║   ██║",
		"██║   ██║",
		"╚██████╔╝",
		" ╚═════╝ ",
	},
	'S': {
		"███████╗",
		"██╔════╝",
		"███████╗",
		"╚════██║",
		"███████║",
		"╚══════╝",
	},
	'A': {
		" █████╗ ",
		"██╔══██╗",
		"███████║",
		"██╔══██║",
		"██║  ██║",
		"╚═╝  ╚═╝",
	},
	'B': {
		"██████╗ ",
		"██╔══██╗",
		"██████╔╝",
		"██╔══██╗",
		"██████╔╝",
		"╚═════╝ ",
	},
	'I': {
		"██╗",
		"██║",
		"██║",
		"██║",
		"██║",
		"╚═╝",
	},
	'W': {
		"██╗    ██╗",
		"██║    ██║",
		"██║ █╗ ██║",
		"██║███╗██║",
		"╚███╔███╔╝",
		" ╚══╝╚══╝ ",
	},
}

// figlet spells word in the shadow face, one string per row.
func figlet(word string) []string {
	rows := make([]string, 6)
	for _, r := range word {
		g := shadow[r]
		for i := range rows {
			rows[i] += g[i]
		}
	}
	for i := range rows {
		rows[i] = strings.TrimRight(rows[i], " ")
	}
	return rows
}

// forge paints a figlet row: the solid blocks in one style and the box drawn
// shadow in another, so the letters sit on a thin line of embers.
func forge(row string, block, ember func(...string) string) string {
	var b strings.Builder
	var run strings.Builder
	solid := false
	flush := func() {
		if run.Len() == 0 {
			return
		}
		if solid {
			b.WriteString(block(run.String()))
		} else {
			b.WriteString(ember(run.String()))
		}
		run.Reset()
	}
	for _, r := range row {
		isSolid := r == '█'
		if r == ' ' {
			// Spaces take no colour; keep them in whichever run is open.
			run.WriteRune(r)
			continue
		}
		if isSolid != solid {
			flush()
			solid = isSolid
		}
		run.WriteRune(r)
	}
	flush()
	return b.String()
}
