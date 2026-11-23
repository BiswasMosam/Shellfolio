package main

// The desk pet's face, ported from its firmware (github.com/BiswasMosam/
// Desktop-pet, src/face.cpp, activity.cpp and media.cpp) so the pet in the
// terminal moves exactly the way the one on the desk does: the same eyes,
// the same easing, the same glances, blinks and fidgets, the same props.
//
// What the pet hears from AMINAL over its cable, this one hears from the
// PC beside it a few minutes late, and only as a mood: code or a game in
// front, music or a film playing, at the desk, or away (see pet.go). The
// parts of the firmware that need more than that are left out: AMINAL's
// listening and speaking faces, the selfie camera, the spectrum bars and
// the jumps a loud moment gives. Visitors can pet it, like the button on
// top of the real one.
//
// One frame is 20 ms, as on the pet, which draws at about 50 a second.

import (
	"fmt"
	"math"
	"math/rand/v2"
	"time"
)

const (
	eyeW   = 36
	eyeH   = 36
	eyeR   = 10 // corner radius
	eyeGap = 16 // space between the eyes

	sleepAfterMs = 30000 // ignored for 30 s, it falls asleep
	nightSleepMs = 10000 // sooner at night
	happyForMs   = 2500  // how long a pet keeps it happy
	frameMs      = 20

	kbX, kbW, kbY = 38, 52, 57 // the keyboard under the paws

	bucketX = 103 // left edge of the popcorn bucket's rim
	bucketY = 46  // its rim
)

const (
	moodNormal = iota
	moodHappy
	moodSleepy
)

const (
	fgNone = iota
	fgLookAround
	fgHop
	fgSquint
	fgDoubleBlink
	fgCount
)

const (
	styleHeadphones = iota
	styleDance
)

const (
	lookPlain = iota
	lookHappy
	lookLove
	lookFocus
)

// petWorld is what the face is told: the twin's version of the firmware's
// `world`.
type petWorld struct {
	act   string // "code", "game" or ""
	media string // "music", "watch" or ""
	away  bool
}

type petPaw struct {
	x         int
	downUntil uint32
}

type petGlyph struct {
	x, y, vx float32
	c        byte
	alive    bool
}

type petNote struct {
	x, y        float32
	alive, pair bool
}

type petFace struct {
	rnd    *rand.Rand
	clock  func() time.Time // for the time it shows while asleep, and night
	now    uint32           // the pet's millis()
	world  petWorld
	typing bool
	screen canvas

	mood      int
	love      bool
	loveUntil uint32

	fidget                  int
	fidgetStart, nextFidget uint32
	fidgetBlinked           bool

	curX, curY, curH, curW float32
	tgtX, tgtY, tgtH, tgtW float32

	nextBlink, blinkEnd, nextLook        uint32
	happyUntil, lastInteraction, lastTap uint32
	tapStreak                            int
	blinking                             bool

	style                   int
	styleUntil              uint32
	shownMedia              string
	lastBeat, nextSynthBeat uint32
	beatNo                  int
	eatStart, nextEat       uint32
	nextDrift               uint32

	lineStart, lineLen, nextKey, nextDart, nextPress uint32
	lineNo                                           int
	lensH                                            float32
	typingUntil                                      uint32

	paws       [2]petPaw
	lastPaw    int
	glyphs     [4]petGlyph
	pressed    int
	pressUntil uint32
	notes      [4]petNote
}

func newPetFace(seed uint64, clock func() time.Time) *petFace {
	p := &petFace{
		rnd:   rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)),
		clock: clock,
		now:   1000,
		curH:  2, // eyes open on boot
		curW:  eyeW,
		tgtH:  eyeH,
		tgtW:  eyeW,
		lensH: 30,
		paws:  [2]petPaw{{x: 51}, {x: 77}},

		lineLen: 1500,
	}
	p.lastInteraction = p.now
	p.nextBlink = p.now + 1500
	p.nextLook = p.now + 1000
	p.nextFidget = p.now + p.random(15000, 40000)
	return p
}

// random is Arduino's random(lo, hi): lo up to but not including hi.
func (p *petFace) random(lo, hi int) uint32 { return uint32(lo + p.rnd.IntN(hi-lo)) }
func (p *petFace) randInt(lo, hi int) int   { return lo + p.rnd.IntN(hi-lo) }

func reached(now, t uint32) bool { return int32(now-t) >= 0 }

// advance runs the pet's loop for n frames, drawing each one, as the
// firmware does. The last one stays on screen.
func (p *petFace) advance(n int) {
	for i := 0; i < n; i++ {
		p.now += frameMs
		p.update()
		p.screen.clear()
		p.draw()
	}
}

// touch is the button on top of the pet: it perks up, and three quick
// pets in a row and it falls for you.
func (p *petFace) touch() {
	now := p.now
	p.wake()
	if now-p.lastTap < 1500 {
		p.tapStreak++
	} else {
		p.tapStreak = 1
	}
	p.lastTap = now
	p.blinking = false
	if p.tapStreak >= 3 {
		p.love, p.loveUntil = true, now+2500
		p.tapStreak = 0
	} else {
		p.mood = moodHappy
		p.happyUntil = now + happyForMs
	}
}

func (p *petFace) wake() {
	p.lastInteraction = p.now
	if p.mood == moodSleepy {
		p.mood = moodNormal
	}
}

func (p *petFace) isNight() bool {
	h := p.clock().In(ist).Hour()
	return h >= 23 || h < 7
}

// ---------- targets ----------

func (p *petFace) startBlink(len uint32) {
	p.blinking = true
	p.blinkEnd = p.now + len
}

func (p *petFace) blinkNaturally(minGap, maxGap int) {
	if !p.blinking && reached(p.now, p.nextBlink) {
		p.startBlink(130)
		p.nextBlink = p.now + p.random(minGap, maxGap)
	}
}

func (p *petFace) idleTargets() {
	now := p.now
	p.tgtH = eyeH
	if p.fidget == fgNone && reached(now, p.nextFidget) {
		p.fidget = p.randInt(1, fgCount)
		p.fidgetStart = now
		p.fidgetBlinked = false
		p.nextFidget = now + p.random(15000, 40000)
	}
	t := now - p.fidgetStart
	switch p.fidget {
	case fgLookAround: // a good look left, then right
		p.tgtY = 0
		p.tgtX = 22
		if t < 700 {
			p.tgtX = -22
		}
		if t > 1500 {
			p.fidget, p.tgtX = fgNone, 0
		}
		return
	case fgHop: // a little bounce
		p.tgtY = 0
		if t < 180 {
			p.tgtY = -9
		}
		if t > 500 {
			p.fidget = fgNone
		}
		return
	case fgSquint: // suspicious
		p.tgtH = 16
		if t > 1000 {
			p.fidget = fgNone
		}
		return
	case fgDoubleBlink:
		if t < 20 {
			p.startBlink(110)
		}
		if t > 300 && !p.fidgetBlinked {
			p.startBlink(110)
			p.fidgetBlinked = true
		}
		if t > 600 {
			p.fidget = fgNone
		}
	}
	// Glance somewhere random every few seconds
	if reached(now, p.nextLook) {
		p.tgtX = float32(p.randInt(-20, 21))
		p.tgtY = float32(p.randInt(-8, 9))
		p.nextLook = now + p.random(1500, 4000)
	}
	p.blinkNaturally(2500, 6000)
}

// Without the PC's spectrum, a song gets headphones or a dance, rerolled
// every so often so a long playlist doesn't look the same all evening.
func (p *petFace) pickStyle() {
	was := p.style
	p.style = p.randInt(0, 2)
	if p.style == was && p.rnd.IntN(4) != 0 {
		p.style = (p.style + 1) % 2
	}
	p.styleUntil = p.now + p.random(25000, 45000)
}

// A steady 115 bpm of its own: the twin hears no beats.
func (p *petFace) tickBeat() {
	if !reached(p.now, p.nextSynthBeat) {
		return
	}
	p.lastBeat = p.now
	p.nextSynthBeat = p.now + 520
	p.beatNo++
	if p.style == styleDance && p.beatNo%2 == 0 && p.world.act == "" {
		p.notesSpawn()
	}
}

func (p *petFace) musicTargets() {
	if reached(p.now, p.styleUntil) {
		p.pickStyle()
	}
	p.tickBeat()
	onBeat := p.now-p.lastBeat < 140
	p.tgtH = eyeH
	switch p.style {
	case styleHeadphones: // eyes shut, nodding along
		p.tgtW, p.tgtX, p.tgtY = 28, 0, -2
		if onBeat {
			p.tgtY = 3
		}
	case styleDance: // side to side, a hop on every beat
		p.tgtX, p.tgtY = -12, 0
		if p.beatNo%2 == 1 {
			p.tgtX = 12
		}
		if onBeat {
			p.tgtY = -8
		}
	}
}

func (p *petFace) watchTargets() {
	now := p.now
	// Eyes up at the screen, following the action a little
	if reached(now, p.nextDrift) {
		p.tgtX = float32(p.randInt(-6, 7))
		p.nextDrift = now + p.random(2500, 6000)
	}
	p.tgtY, p.tgtH = -8, 34
	if reached(now, p.nextEat) {
		p.eatStart = now
		p.nextEat = now + p.random(3500, 7000)
	}
	eatT := now - p.eatStart
	if p.eatStart != 0 && eatT > 450 && eatT < 1150 {
		p.tgtH = 28 // chewing, content
	}
	p.blinkNaturally(3000, 7000)
}

// Headphones on over whatever it's doing: a nod on every beat
func (p *petFace) nodToMusic() {
	if p.world.media != "music" {
		return
	}
	p.tickBeat()
	if p.now-p.lastBeat < 140 {
		p.tgtY += 2
	}
}

func (p *petFace) codeTargets() {
	now := p.now
	// The twin can't see the keys, so it types in bursts and reads in
	// between, the way an evening of coding looks.
	if reached(now, p.typingUntil) {
		p.typing = !p.typing
		if p.typing {
			p.typingUntil = now + p.random(3000, 9000)
		} else {
			p.typingUntil = now + p.random(4000, 12000)
		}
	}
	p.tgtW, p.tgtH = 32, 24
	if p.typing {
		// Eyes on the keys, paws going
		p.tgtX, p.tgtY = 0, -4
		if reached(now, p.nextKey) {
			p.pawTap()
			p.nextKey = now + p.random(70, 190)
		}
	} else {
		// Reading: along a line, a quick hop back, and on down the page
		t := now - p.lineStart
		if t > p.lineLen {
			p.lineStart = now
			p.lineLen = p.random(1200, 2400)
			p.lineNo++
			t = 0
		}
		p.tgtX = float32(-7 + 14*int32(t)/int32(p.lineLen))
		p.tgtY = float32(-7 + p.lineNo%3)
	}
	p.blinkNaturally(3500, 8000)
	p.nodToMusic()
}

func (p *petFace) gameTargets() {
	now := p.now
	// Darting after the action, barely blinking, thumbs never still
	if reached(now, p.nextDart) {
		p.tgtX = float32(p.randInt(-12, 13))
		p.nextDart = now + p.random(250, 900)
	}
	p.tgtW, p.tgtY, p.tgtH = 34, -8, 22
	if reached(now, p.nextPress) {
		p.pressed = p.randInt(1, 9)
		p.pressUntil = now + 120
		p.nextPress = now + p.random(90, 280)
	}
	p.blinkNaturally(5000, 10000)
	p.nodToMusic()
}

func (p *petFace) update() {
	now := p.now
	if p.love && reached(now, p.loveUntil) {
		p.love = false
	}
	if p.mood == moodHappy && reached(now, p.happyUntil) {
		p.mood = moodNormal
	}

	// The pet on the desk dozes off whenever nothing is going on. The twin
	// stays up while he is at the desk, so a visitor sees it awake when he
	// is there and asleep when he isn't.
	if !p.world.away {
		p.wake()
	}

	if p.world.media != p.shownMedia {
		p.shownMedia = p.world.media
		if p.shownMedia == "music" {
			p.pickStyle()
		}
		if p.shownMedia == "watch" {
			p.eatStart = 0
			p.nextEat = now + p.random(1500, 3000)
			p.nextDrift = now
		}
	}

	sleepAfter := uint32(sleepAfterMs)
	if p.world.away || p.isNight() {
		sleepAfter = nightSleepMs
	}
	if p.mood == moodNormal && now-p.lastInteraction > sleepAfter {
		p.mood = moodSleepy
	}

	p.tgtW = eyeW
	switch {
	case p.love:
		p.tgtX, p.tgtY, p.tgtH = 0, 0, eyeH
	case p.mood == moodHappy:
		p.tgtX, p.tgtY, p.tgtH = 0, -4, eyeH
	case p.world.act == "code":
		p.codeTargets()
	case p.world.act == "game":
		p.gameTargets()
	case p.world.media == "music":
		p.musicTargets()
	case p.world.media == "watch":
		p.watchTargets()
	case p.mood == moodSleepy:
		p.tgtX, p.tgtY, p.tgtH = 0, 8, 6
	default:
		p.idleTargets()
	}

	p.lensH += (p.tgtH + 6 - p.lensH) * 0.25 // before a blink shuts the target

	if p.blinking && reached(now, p.blinkEnd) {
		p.blinking = false
	}
	if p.blinking {
		p.tgtH = 2
	}

	// Ease toward the target. Blinks close fast, everything else glides.
	speed := float32(0.25)
	if p.blinking || p.tgtH < p.curH-10 {
		speed = 0.6
	}
	p.curX += (p.tgtX - p.curX) * 0.25
	p.curY += (p.tgtY - p.curY) * 0.25
	p.curW += (p.tgtW - p.curW) * 0.25
	p.curH += (p.tgtH - p.curH) * speed
}

// ---------- drawing ----------

func (p *petFace) actShown() bool   { return p.world.act != "" }
func (p *petFace) mediaShown() bool { return p.world.media != "" && p.world.act == "" }

func (p *petFace) currentLook() int {
	switch {
	case p.love:
		return lookLove
	case p.mood == moodHappy:
		return lookHappy
	case p.actShown():
		// Game face: lids slanted in
		if p.world.act == "game" {
			return lookFocus
		}
		return lookPlain
	case p.mediaShown() && p.world.media == "music":
		// Lost in it, with a peek around every nine seconds or so
		if p.style == styleHeadphones {
			if (p.now/1500)%6 == 0 {
				return lookPlain
			}
			return lookHappy
		}
		return lookHappy
	}
	return lookPlain
}

func (p *petFace) drawHeart(cx, cy, s int) {
	c := &p.screen
	r := s / 4
	c.fillCircle(cx-r, cy-r/2, r, true)
	c.fillCircle(cx+r, cy-r/2, r, true)
	c.fillTriangle(cx-2*r, cy-r/2, cx+2*r, cy-r/2, cx, cy+2*r, true)
}

// side is -1 for the left eye, +1 for the right
func (p *petFace) drawEye(cx, cy, w, h, side, look int) {
	c := &p.screen
	if look == lookLove {
		beat := 0
		if (p.now/150)%4 == 0 { // a heartbeat
			beat = 4
		}
		p.drawHeart(cx, cy-2, 34+beat)
		return
	}
	h = max(h, 2)
	r := min(eyeR, h/2)
	x0, y0 := cx-w/2, cy-h/2
	c.fillRoundRect(x0, y0, w, h, r, true)

	switch look {
	case lookHappy:
		// Cut a curve out of the bottom so the eye becomes a happy "^".
		c.fillCircle(cx, cy+eyeH/2+w/4, (w*6)/10, false)
	case lookFocus:
		// A lid slanting down across the whole eye toward the nose
		inner, outer := x0-1, x0+w
		if side < 0 {
			inner, outer = x0+w, x0-1
		}
		c.fillTriangle(outer, y0-1, inner, y0-1, inner, y0+h/3, false)
	}
}

func (p *petFace) draw() {
	c := &p.screen
	now := p.now
	media := p.mediaShown()
	plain := !p.love && p.mood != moodHappy

	look := p.currentLook()
	cy := screenH/2 + int(p.curY)
	w, h := int(p.curW), int(p.curH)
	lx := screenW/2 - eyeGap/2 - eyeW/2 + int(p.curX)
	rx := screenW/2 + eyeGap/2 + eyeW/2 + int(p.curX)

	act := p.actShown()
	if act && p.world.act == "code" { // under the eyes
		p.keyboardDraw()
		p.pawsDraw()
		p.glyphsDraw()
	}

	p.drawEye(lx, cy, w, h, -1, look)
	p.drawEye(rx, cy, w, h, +1, look)

	switch {
	case act:
		if p.world.act == "code" {
			p.glassesDraw(lx, rx, cy, w, int(p.lensH))
		} else {
			p.controllerDraw(int(p.curX) / 4)
		}
		if p.world.media == "music" {
			p.headphonesDraw(lx, rx, cy, w)
		}
	case media && p.world.media == "music":
		if p.style == styleHeadphones {
			p.headphonesDraw(lx, rx, cy, w)
		} else {
			p.notesDraw()
		}
	case media && p.world.media == "watch":
		p.bucketDraw()
		mouthX, mouthY := screenW/2+int(p.curX), cy+h/2+5
		eatT := now - p.eatStart
		if p.eatStart != 0 && plain {
			p.kernelDraw(eatT, mouthX, mouthY)
			if eatT >= 450 && eatT < 1150 {
				hh := 2
				if (now/120)%2 == 1 {
					hh = 5
				}
				c.fillRoundRect(mouthX-5, mouthY, 10, hh, 2, true)
			}
		}
	}

	if !p.love && p.mood == moodSleepy {
		// Little floating z that bobs up and down
		bob := int((now / 400) % 3)
		c.print(112, 8-bob, "z")
		c.print(118, 2-bob/2, "z")
		// and the time, small, for anyone glancing over at night
		t := p.clock().In(ist)
		hr := t.Hour() % 12
		if hr == 0 {
			hr = 12
		}
		c.print(0, 0, fmt.Sprintf("%d:%02d", hr, t.Minute()))
	}
}

// ---------- coding and gaming props ----------

// A frame around each eye and a bridge between them. The height is the
// eye's open height, so a blink closes the eye behind the glasses, not
// the glasses with it.
func (p *petFace) glassesDraw(lx, rx, cy, ew, lensH int) {
	c := &p.screen
	for _, cx := range []int{lx, rx} {
		x, y := cx-ew/2-3, cy-lensH/2
		c.roundRect(x, y, ew+6, lensH, 6, true)
		c.roundRect(x+1, y+1, ew+4, lensH-2, 5, true)
	}
	from, to := lx+ew/2+3, rx-ew/2-3
	y := cy - lensH/2 + 6
	c.hline(from, y, to-from, true)
	c.hline(from+2, y-1, to-from-4, true)
}

func (p *petFace) glyphSpawn(side int) {
	const code = "{}();<>=/*#01"
	for i := range p.glyphs {
		g := &p.glyphs[i]
		if g.alive {
			continue
		}
		g.x = kbX + kbW
		if side < 0 {
			g.x = kbX - 6
		}
		g.y = kbY - 6
		g.vx = float32(side) * 1.4
		g.c = code[p.rnd.IntN(len(code))]
		g.alive = true
		return
	}
}

func (p *petFace) pawTap() {
	i := p.lastPaw // mostly one paw then the other
	if p.rnd.IntN(10) < 7 {
		i = 1 - p.lastPaw
	}
	p.lastPaw = i
	p.paws[i].downUntil = p.now + 80
	if p.rnd.IntN(3) == 0 {
		side := -1
		if i == 1 {
			side = 1
		}
		p.glyphSpawn(side)
	}
}

func (p *petFace) keyboardDraw() {
	c := &p.screen
	c.fillRoundRect(kbX, kbY, kbW, 7, 2, true)
	for x := kbX + 3; x < kbX+kbW-3; x += 4 {
		c.hline(x, kbY+2, 2, false)
		c.hline(x+2, kbY+4, 2, false)
	}
}

// Resting on the keys while it reads; up and down while it types
func (p *petFace) pawsDraw() {
	c := &p.screen
	for _, paw := range p.paws {
		down := !p.typing || !reached(p.now, paw.downUntil)
		y := 47
		if down {
			y = 52
		}
		c.fillRoundRect(paw.x-6, y-1, 12, 8, 3, false) // keeps it off the keys
		c.fillRoundRect(paw.x-5, y, 10, 6, 3, true)
		c.vline(paw.x-2, y+4, 2, false) // toes
		c.vline(paw.x+1, y+4, 2, false)
	}
}

func (p *petFace) glyphsDraw() {
	for i := range p.glyphs {
		g := &p.glyphs[i]
		if !g.alive {
			continue
		}
		g.y -= 0.5 // out to the side, under the glasses
		g.x += g.vx
		if g.y < 42 {
			g.alive = false
			continue
		}
		p.screen.char(int(g.x), int(g.y), g.c)
	}
}

// dx leans the controller with the eyes
func (p *petFace) controllerDraw(dx int) {
	c := &p.screen
	cx, top := screenW/2+dx, 47
	c.fillRoundRect(cx-22, top, 44, 12, 5, true)
	c.fillCircle(cx-16, top+12, 6, true) // the grips
	c.fillCircle(cx+16, top+12, 6, true)

	pressed := 0
	if !reached(p.now, p.pressUntil) {
		pressed = p.pressed
	}
	// The d-pad, nudged toward the direction held
	px, py := cx-14, top+6
	switch pressed {
	case 5:
		px--
	case 6:
		px++
	case 7:
		py--
	case 8:
		py++
	}
	c.fillRect(px-4, py-1, 9, 3, false)
	c.fillRect(px-1, py-4, 3, 9, false)

	// Four face buttons in a diamond; the one pressed swells
	at := [4][2]int{{0, -3}, {3, 0}, {0, 3}, {-3, 0}}
	for i, a := range at {
		r := 1
		if pressed == i+1 {
			r = 2
		}
		c.fillCircle(cx+14+a[0], top+6+a[1], r, false)
	}
}

// ---------- music and film props ----------

// A band arching over the top of the screen from cup to cup, moving with
// the head so a nod takes the headphones with it
func (p *petFace) headphonesDraw(lx, rx, cy, ew int) {
	c := &p.screen
	leftCup, rightCup := lx-ew/2-14, rx+ew/2+4
	cupTop := cy - 12
	c.fillRoundRect(leftCup, cupTop, 10, 24, 4, true)
	c.fillRoundRect(rightCup, cupTop, 10, 24, 4, true)

	// Half an ellipse from one cup to the other, two pixels thick
	cx := float32(leftCup+5+rightCup+5) / 2
	rx2 := float32(rightCup-leftCup) / 2
	ry := float32(max(cupTop-1, 4))
	var px, py int
	for i := 0; i <= 20; i++ {
		a := float32(math.Pi * float64(i) / 20)
		x := int(cx - cosf(a)*rx2)
		y := int(float32(cupTop) - sinf(a)*ry)
		if i > 0 {
			c.line(px, py, x, y, true)
			c.line(px, py+1, x, y+1, true)
		}
		px, py = x, y
	}
}

func (p *petFace) notesSpawn() {
	for i := range p.notes {
		n := &p.notes[i]
		if n.alive {
			continue
		}
		if p.rnd.IntN(2) == 1 {
			n.x = float32(p.randInt(4, 18))
		} else {
			n.x = float32(p.randInt(108, 120))
		}
		n.y = 56
		n.pair = p.rnd.IntN(3) == 0
		n.alive = true
		return
	}
}

func (p *petFace) notesDraw() {
	c := &p.screen
	for i := range p.notes {
		n := &p.notes[i]
		if !n.alive {
			continue
		}
		n.y -= 0.7
		n.x += sinf(n.y/6) * 0.4
		if n.y < 4 {
			n.alive = false
			continue
		}
		x, y := int(n.x), int(n.y)
		c.fillCircle(x, y, 2, true)
		c.vline(x+2, y-7, 7, true)
		if n.pair { // a beamed pair
			c.fillCircle(x+6, y-1, 2, true)
			c.vline(x+8, y-8, 7, true)
			c.line(x+2, y-7, x+8, y-8, true)
			c.line(x+2, y-6, x+8, y-7, true)
		} else {
			c.line(x+2, y-7, x+5, y-4, true)
		}
	}
}

func (p *petFace) bucketDraw() {
	c := &p.screen
	// A striped bucket, narrower at the bottom
	for y := bucketY; y < screenH; y++ {
		w := 23 - (y-bucketY)*6/17
		c.hline(bucketX+(23-w)/2, y, w, true)
	}
	for s := 0; s < 3; s++ {
		top := bucketX + 5 + s*6
		c.line(top, bucketY+2, top+(s-1), screenH-1, false)
	}
	// Heaped over the rim
	heap := [][2]int{{2, -2}, {7, -4}, {12, -3}, {17, -4}, {21, -2}, {9, -8}, {15, -7}}
	for _, k := range heap {
		c.fillCircle(bucketX+k[0], bucketY+k[1], 3, true)
		c.pixel(bucketX+k[0], bucketY+k[1], false) // a crease
	}
}

// One kernel on its way from the bucket to the mouth, in the first 450 ms
func (p *petFace) kernelDraw(eatT uint32, mouthX, mouthY int) {
	if eatT >= 450 {
		return
	}
	t := float32(eatT) / 450
	fromX, fromY := bucketX+11, bucketY-10
	x := int(float32(fromX) + float32(mouthX-fromX)*t)
	y := int(float32(fromY) + float32(mouthY-fromY)*t - sinf(math.Pi*t)*16)
	p.screen.fillCircle(x, y, 2, true)
}

func cosf(a float32) float32 { return float32(math.Cos(float64(a))) }
func sinf(a float32) float32 { return float32(math.Sin(float64(a))) }
