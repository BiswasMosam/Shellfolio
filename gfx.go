package main

// A 128x64 one-bit screen with the drawing calls the desk pet's firmware
// uses, ported from Adafruit GFX and Adafruit SSD1306 so that a shape drawn
// here lands on the same pixels it does on the pet's OLED. Only what the
// face needs is here.
//
// The primitives are adapted from Adafruit GFX Library,
// Copyright (c) 2012 Adafruit Industries, BSD licence.

import "strings"

const (
	screenW = 128
	screenH = 64
)

type canvas [screenH][screenW]bool

func (c *canvas) clear() { *c = canvas{} }

func (c *canvas) pixel(x, y int, on bool) {
	if x < 0 || y < 0 || x >= screenW || y >= screenH {
		return
	}
	c[y][x] = on
}

// hline and vline clip like the SSD1306 driver: a length of zero or less
// draws nothing.
func (c *canvas) hline(x, y, w int, on bool) {
	if y < 0 || y >= screenH {
		return
	}
	if x < 0 {
		w += x
		x = 0
	}
	if x+w > screenW {
		w = screenW - x
	}
	for ; w > 0; w-- {
		c[y][x] = on
		x++
	}
}

func (c *canvas) vline(x, y, h int, on bool) {
	if x < 0 || x >= screenW {
		return
	}
	if y < 0 {
		h += y
		y = 0
	}
	if y+h > screenH {
		h = screenH - y
	}
	for ; h > 0; h-- {
		c[y][x] = on
		y++
	}
}

func (c *canvas) fillRect(x, y, w, h int, on bool) {
	for i := x; i < x+w; i++ {
		c.vline(i, y, h, on)
	}
}

func (c *canvas) line(x0, y0, x1, y1 int, on bool) {
	if x0 == x1 {
		if y0 > y1 {
			y0, y1 = y1, y0
		}
		c.vline(x0, y0, y1-y0+1, on)
		return
	}
	if y0 == y1 {
		if x0 > x1 {
			x0, x1 = x1, x0
		}
		c.hline(x0, y0, x1-x0+1, on)
		return
	}
	steep := iabs(y1-y0) > iabs(x1-x0)
	if steep {
		x0, y0 = y0, x0
		x1, y1 = y1, x1
	}
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}
	dx, dy := x1-x0, iabs(y1-y0)
	err := dx / 2
	ystep := -1
	if y0 < y1 {
		ystep = 1
	}
	for ; x0 <= x1; x0++ {
		if steep {
			c.pixel(y0, x0, on)
		} else {
			c.pixel(x0, y0, on)
		}
		err -= dy
		if err < 0 {
			y0 += ystep
			err += dx
		}
	}
}

func (c *canvas) circleCorners(x0, y0, r int, corners uint8, on bool) {
	f, ddx, ddy, x, y := 1-r, 1, -2*r, 0, r
	for x < y {
		if f >= 0 {
			y--
			ddy += 2
			f += ddy
		}
		x++
		ddx += 2
		f += ddx
		if corners&4 != 0 {
			c.pixel(x0+x, y0+y, on)
			c.pixel(x0+y, y0+x, on)
		}
		if corners&2 != 0 {
			c.pixel(x0+x, y0-y, on)
			c.pixel(x0+y, y0-x, on)
		}
		if corners&8 != 0 {
			c.pixel(x0-y, y0+x, on)
			c.pixel(x0-x, y0+y, on)
		}
		if corners&1 != 0 {
			c.pixel(x0-y, y0-x, on)
			c.pixel(x0-x, y0-y, on)
		}
	}
}

func (c *canvas) fillCircle(x0, y0, r int, on bool) {
	c.vline(x0, y0-r, 2*r+1, on)
	c.fillHalves(x0, y0, r, 3, 0, on)
}

func (c *canvas) fillHalves(x0, y0, r int, corners uint8, delta int, on bool) {
	f, ddx, ddy, x, y := 1-r, 1, -2*r, 0, r
	px, py := x, y
	delta++
	for x < y {
		if f >= 0 {
			y--
			ddy += 2
			f += ddy
		}
		x++
		ddx += 2
		f += ddx
		if x < y+1 {
			if corners&1 != 0 {
				c.vline(x0+x, y0-y, 2*y+delta, on)
			}
			if corners&2 != 0 {
				c.vline(x0-x, y0-y, 2*y+delta, on)
			}
		}
		if y != py {
			if corners&1 != 0 {
				c.vline(x0+py, y0-px, 2*px+delta, on)
			}
			if corners&2 != 0 {
				c.vline(x0-py, y0-px, 2*px+delta, on)
			}
			py = y
		}
		px = x
	}
}

func (c *canvas) roundRect(x, y, w, h, r int, on bool) {
	r = min(r, min(w, h)/2)
	c.hline(x+r, y, w-2*r, on)
	c.hline(x+r, y+h-1, w-2*r, on)
	c.vline(x, y+r, h-2*r, on)
	c.vline(x+w-1, y+r, h-2*r, on)
	c.circleCorners(x+r, y+r, r, 1, on)
	c.circleCorners(x+w-r-1, y+r, r, 2, on)
	c.circleCorners(x+w-r-1, y+h-r-1, r, 4, on)
	c.circleCorners(x+r, y+h-r-1, r, 8, on)
}

func (c *canvas) fillRoundRect(x, y, w, h, r int, on bool) {
	r = min(r, min(w, h)/2)
	c.fillRect(x+r, y, w-2*r, h, on)
	c.fillHalves(x+w-r-1, y+r, r, 1, h-2*r-1, on)
	c.fillHalves(x+r, y+r, r, 2, h-2*r-1, on)
}

func (c *canvas) fillTriangle(x0, y0, x1, y1, x2, y2 int, on bool) {
	if y0 > y1 {
		y0, y1 = y1, y0
		x0, x1 = x1, x0
	}
	if y1 > y2 {
		y2, y1 = y1, y2
		x2, x1 = x1, x2
	}
	if y0 > y1 {
		y0, y1 = y1, y0
		x0, x1 = x1, x0
	}
	if y0 == y2 {
		a, b := x0, x0
		if x1 < a {
			a = x1
		} else if x1 > b {
			b = x1
		}
		if x2 < a {
			a = x2
		} else if x2 > b {
			b = x2
		}
		c.hline(a, y0, b-a+1, on)
		return
	}
	dx01, dy01, dx02, dy02, dx12, dy12 := x1-x0, y1-y0, x2-x0, y2-y0, x2-x1, y2-y1
	sa, sb := 0, 0
	last := y1 - 1
	if y1 == y2 {
		last = y1
	}
	y := y0
	for ; y <= last; y++ {
		a, b := x0+sa/dy01, x0+sb/dy02
		sa += dx01
		sb += dx02
		if a > b {
			a, b = b, a
		}
		c.hline(a, y, b-a+1, on)
	}
	sa, sb = dx12*(y-y1), dx02*(y-y0)
	for ; y <= y2; y++ {
		a, b := x1+sa/dy12, x0+sb/dy02
		sa += dx12
		sb += dx02
		if a > b {
			a, b = b, a
		}
		c.hline(a, y, b-a+1, on)
	}
}

// The classic 5x7 font, only the characters the face ever prints: the
// sleeping z, the clock, and the code that flies off the keyboard. Each
// column is a byte, lowest bit at the top.
var glyphs5x7 = map[byte][5]byte{
	'z': {0x44, 0x64, 0x54, 0x4C, 0x44},
	'0': {0x3E, 0x51, 0x49, 0x45, 0x3E},
	'1': {0x00, 0x42, 0x7F, 0x40, 0x00},
	'2': {0x72, 0x49, 0x49, 0x49, 0x46},
	'3': {0x21, 0x41, 0x49, 0x4D, 0x33},
	'4': {0x18, 0x14, 0x12, 0x7F, 0x10},
	'5': {0x27, 0x45, 0x45, 0x45, 0x39},
	'6': {0x3C, 0x4A, 0x49, 0x49, 0x31},
	'7': {0x41, 0x21, 0x11, 0x09, 0x07},
	'8': {0x36, 0x49, 0x49, 0x49, 0x36},
	'9': {0x46, 0x49, 0x49, 0x29, 0x1E},
	':': {0x00, 0x00, 0x14, 0x00, 0x00},
	'{': {0x00, 0x08, 0x36, 0x41, 0x00},
	'}': {0x00, 0x41, 0x36, 0x08, 0x00},
	'(': {0x00, 0x1C, 0x22, 0x41, 0x00},
	')': {0x00, 0x41, 0x22, 0x1C, 0x00},
	';': {0x00, 0x40, 0x34, 0x00, 0x00},
	'<': {0x00, 0x08, 0x14, 0x22, 0x41},
	'>': {0x00, 0x41, 0x22, 0x14, 0x08},
	'=': {0x14, 0x14, 0x14, 0x14, 0x14},
	'/': {0x20, 0x10, 0x08, 0x04, 0x02},
	'*': {0x2A, 0x1C, 0x7F, 0x1C, 0x2A},
	'#': {0x14, 0x7F, 0x14, 0x7F, 0x14},
}

// char draws one character with a clear background, like the firmware's
// print with the text and background colours the same.
func (c *canvas) char(x, y int, ch byte) {
	cols, ok := glyphs5x7[ch]
	if !ok {
		return
	}
	for i, col := range cols {
		for j := 0; j < 8; j, col = j+1, col>>1 {
			if col&1 != 0 {
				c.pixel(x+i, y+j, true)
			}
		}
	}
}

func (c *canvas) print(x, y int, s string) {
	for i := 0; i < len(s); i++ {
		c.char(x+6*i, y, s[i])
	}
}

// braille draws the screen with Unicode braille, one character for every
// two by four pixels, so the whole 128x64 screen is 64 columns by 16 rows
// and nothing is lost. Empty cells are spaces, so the text copies cleanly.
func (c *canvas) braille() []string {
	return c.brailleOf(func(x, y int) bool { return c[y][x] }, screenW, screenH)
}

// brailleSmall is the same at half size, 32 columns by 8 rows, for narrow
// terminals. A pixel is on if any of the four it stands for is, so a one
// pixel line survives the shrink.
func (c *canvas) brailleSmall() []string {
	on := func(x, y int) bool {
		return c[2*y][2*x] || c[2*y][2*x+1] || c[2*y+1][2*x] || c[2*y+1][2*x+1]
	}
	return c.brailleOf(on, screenW/2, screenH/2)
}

// Bit for each dot of a braille cell, by column then row.
var brailleDots = [2][4]rune{{0x01, 0x02, 0x04, 0x40}, {0x08, 0x10, 0x20, 0x80}}

func (c *canvas) brailleOf(on func(x, y int) bool, w, h int) []string {
	lines := make([]string, 0, h/4)
	for cy := 0; cy < h; cy += 4 {
		var b strings.Builder
		for cx := 0; cx < w; cx += 2 {
			var bits rune
			for dx := 0; dx < 2; dx++ {
				for dy := 0; dy < 4; dy++ {
					if on(cx+dx, cy+dy) {
						bits |= brailleDots[dx][dy]
					}
				}
			}
			if bits == 0 {
				b.WriteByte(' ')
			} else {
				b.WriteRune(0x2800 + bits)
			}
		}
		lines = append(lines, b.String())
	}
	return lines
}

func iabs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
